import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { App } from "./App";

describe("App", () => {
  it("renders the AI-For-OJ shell", () => {
    render(<App />);

    expect(screen.getByText("AI-For-OJ")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Experiment Console" })).toBeInTheDocument();
  });
});
