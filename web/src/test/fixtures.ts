export const experimentOptionsFixture = {
  default_model: 'mock-cpp17',
  prompts: [
    { name: 'default', label: 'default' },
    { name: 'cpp17_minimal', label: 'cpp17_minimal' },
    { name: 'strict_cpp17', label: 'strict_cpp17' },
  ],
  agents: [
    { name: 'direct_codegen', label: 'direct_codegen' },
    { name: 'direct_codegen_repair', label: 'direct_codegen_repair' },
    { name: 'analyze_then_codegen', label: 'analyze_then_codegen' },
  ],
};
