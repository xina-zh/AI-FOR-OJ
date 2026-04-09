package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/model"
)

const (
	sourceFileName     = "main.cpp"
	binaryFileName     = "main"
	dockerInfraExit    = 125
	defaultTimeLimitMS = 1000
)

var sandboxSequence uint64

type DockerSandbox struct {
	logger           *slog.Logger
	workDir          string
	dockerImage      string
	compileTimeout   time.Duration
	runTimeoutBuffer time.Duration
	compileMemoryMB  int
}

func NewDockerSandbox(cfg config.SandboxConfig, logger *slog.Logger) (*DockerSandbox, error) {
	if cfg.WorkDir == "" {
		return nil, errors.New("sandbox work dir is required")
	}
	if cfg.DockerImage == "" {
		return nil, errors.New("sandbox docker image is required")
	}

	if err := os.MkdirAll(cfg.WorkDir, 0o755); err != nil {
		return nil, fmt.Errorf("create sandbox work dir: %w", err)
	}

	return &DockerSandbox{
		logger:           logger,
		workDir:          cfg.WorkDir,
		dockerImage:      cfg.DockerImage,
		compileTimeout:   cfg.CompileTimeout,
		runTimeoutBuffer: cfg.RunTimeoutBuffer,
		compileMemoryMB:  cfg.CompileMemoryMB,
	}, nil
}

func (s *DockerSandbox) Compile(ctx context.Context, req CompileRequest) (CompileResult, error) {
	if req.Language != model.LanguageCPP17 {
		return CompileResult{}, fmt.Errorf("unsupported sandbox language: %s", req.Language)
	}

	artifactDir, err := os.MkdirTemp(s.workDir, "artifact-*")
	if err != nil {
		return CompileResult{}, fmt.Errorf("create artifact dir: %w", err)
	}

	sourcePath := filepath.Join(artifactDir, sourceFileName)
	if err := os.WriteFile(sourcePath, []byte(req.SourceCode), 0o644); err != nil {
		_ = os.RemoveAll(artifactDir)
		return CompileResult{}, fmt.Errorf("write source file: %w", err)
	}

	containerName := s.containerName("compile")
	compileCtx, cancel := context.WithTimeout(ctx, s.compileTimeout)
	defer cancel()

	command := []string{
		"run",
		"--rm",
		"--name", containerName,
		"--network", "none",
		"--cpus", "1",
		"--memory", fmt.Sprintf("%dm", s.compileMemoryMB),
		"--pids-limit", "128",
		"-v", fmt.Sprintf("%s:/workspace", artifactDir),
		"-w", "/workspace",
		s.dockerImage,
		"sh", "-lc",
		fmt.Sprintf("g++ -std=c++17 -O2 -pipe %s -o %s", sourceFileName, binaryFileName),
	}

	stdout, stderr, exitCode, timedOut, err := s.runDockerCommand(compileCtx, containerName, nil, command...)
	if err != nil {
		_ = os.RemoveAll(artifactDir)
		return CompileResult{}, fmt.Errorf("run compiler container: %w", err)
	}

	if timedOut {
		_ = os.RemoveAll(artifactDir)
		return CompileResult{
			Success:      false,
			Stderr:       stderr,
			ExitCode:     exitCode,
			ErrorMessage: "compile timed out",
		}, nil
	}

	if exitCode == dockerInfraExit {
		_ = os.RemoveAll(artifactDir)
		return CompileResult{}, s.infrastructureError("compile", stderr)
	}

	if exitCode != 0 {
		_ = os.RemoveAll(artifactDir)
		return CompileResult{
			Success:      false,
			Stderr:       stderr,
			ExitCode:     exitCode,
			ErrorMessage: strings.TrimSpace(stderr),
		}, nil
	}

	if s.logger != nil {
		s.logger.Info("sandbox compile completed",
			"artifact_dir", artifactDir,
			"stdout", strings.TrimSpace(stdout),
		)
	}

	return CompileResult{
		Success:    true,
		ArtifactID: artifactDir,
		Stderr:     stderr,
		ExitCode:   0,
	}, nil
}

func (s *DockerSandbox) Run(ctx context.Context, req RunRequest) (RunResult, error) {
	if req.Language != model.LanguageCPP17 {
		return RunResult{}, fmt.Errorf("unsupported sandbox language: %s", req.Language)
	}

	containerName := s.containerName("run")
	runTimeout := s.runTimeout(req.TimeLimitMS)
	runCtx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	startedAt := time.Now()
	command := []string{
		"run",
		"--rm",
		"-i",
		"--name", containerName,
		"--network", "none",
		"--cpus", "1",
		"--memory", s.memoryLimitArg(req.MemoryLimitMB),
		"--pids-limit", "64",
		"-v", fmt.Sprintf("%s:/workspace:ro", req.ArtifactID),
		"-w", "/workspace",
		s.dockerImage,
		fmt.Sprintf("/workspace/%s", binaryFileName),
	}

	stdout, stderr, exitCode, timedOut, err := s.runDockerCommand(runCtx, containerName, strings.NewReader(req.Input), command...)
	if err != nil {
		return RunResult{}, fmt.Errorf("run program container: %w", err)
	}

	runtimeMS := int(time.Since(startedAt).Milliseconds())
	if timedOut {
		return RunResult{
			Stdout:    stdout,
			Stderr:    stderr,
			ExitCode:  exitCode,
			RuntimeMS: runtimeMS,
			MemoryKB:  0,
			TimedOut:  true,
		}, nil
	}

	if exitCode == dockerInfraExit {
		return RunResult{}, s.infrastructureError("run", stderr)
	}

	return RunResult{
		Stdout:       stdout,
		Stderr:       stderr,
		ExitCode:     exitCode,
		RuntimeMS:    runtimeMS,
		MemoryKB:     0,
		RuntimeError: exitCode != 0,
		ErrorMessage: strings.TrimSpace(stderr),
	}, nil
}

func (s *DockerSandbox) Cleanup(_ context.Context, artifactID string) error {
	if artifactID == "" {
		return nil
	}
	return os.RemoveAll(artifactID)
}

func (s *DockerSandbox) runDockerCommand(
	ctx context.Context,
	containerName string,
	stdin io.Reader,
	args ...string,
) (stdout string, stderr string, exitCode int, timedOut bool, err error) {
	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	if stdin != nil {
		cmd.Stdin = stdin
	}

	runErr := cmd.Run()
	stdout = stdoutBuffer.String()
	stderr = stderrBuffer.String()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		s.forceRemoveContainer(containerName)
		return stdout, stderr, -1, true, nil
	}

	if runErr == nil {
		return stdout, stderr, 0, false, nil
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		return stdout, stderr, exitErr.ExitCode(), false, nil
	}

	return stdout, stderr, 0, false, runErr
}

func (s *DockerSandbox) forceRemoveContainer(containerName string) {
	removeCmd := exec.Command("docker", "rm", "-f", containerName)
	output, err := removeCmd.CombinedOutput()
	if err != nil && s.logger != nil {
		s.logger.Warn("force remove container failed",
			"container_name", containerName,
			"error", err,
			"output", strings.TrimSpace(string(output)),
		)
	}
}

func (s *DockerSandbox) runTimeout(timeLimitMS int) time.Duration {
	if timeLimitMS <= 0 {
		timeLimitMS = defaultTimeLimitMS
	}
	return time.Duration(timeLimitMS)*time.Millisecond + s.runTimeoutBuffer
}

func (s *DockerSandbox) memoryLimitArg(memoryLimitMB int) string {
	if memoryLimitMB <= 0 {
		memoryLimitMB = 256
	}
	return strconv.Itoa(memoryLimitMB) + "m"
}

func (s *DockerSandbox) containerName(stage string) string {
	seq := atomic.AddUint64(&sandboxSequence, 1)
	return fmt.Sprintf("ai-for-oj-%s-%d-%d", stage, time.Now().UnixNano(), seq)
}

func (s *DockerSandbox) infrastructureError(stage, stderr string) error {
	message := strings.TrimSpace(stderr)
	switch {
	case containsAny(message, "No such image:", "Unable to find image"):
		return fmt.Errorf(
			"docker sandbox %s failed: required image %q is not available locally. This is a runtime environment issue; pull the image first. docker stderr: %s",
			stage,
			s.dockerImage,
			message,
		)
	case containsAny(message, "permission denied while trying to connect to the Docker daemon socket"):
		return fmt.Errorf(
			"docker sandbox %s failed: cannot access the Docker daemon. This is a runtime environment issue. docker stderr: %s",
			stage,
			message,
		)
	default:
		return fmt.Errorf("docker sandbox %s infrastructure error: %s", stage, message)
	}
}

func containsAny(value string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}
