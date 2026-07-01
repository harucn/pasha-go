import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { RunTestSession } from "../wailsjs/go/main/App";

vi.mock("../wailsjs/go/main/App", () => ({
	RunTestSession: vi.fn(() => Promise.resolve()),
}));

import App from "./App";

describe("App", () => {
	it("renders the initial prompt", () => {
		render(<App />);
		expect(screen.getByText(/press the button/i)).toBeInTheDocument();
	});

	it("shows completion message after running a test session", async () => {
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(
			await screen.findByText(/check ~\/Desktop\/pasha-tracer\.pdf/i),
		).toBeInTheDocument();
	});

	it("renders a Repeat Count input with default value 10", () => {
		render(<App />);
		const input = screen.getByLabelText(/repeat count/i) as HTMLInputElement;
		expect(input).toBeInTheDocument();
		expect(input.type).toBe("number");
		expect(input.value).toBe("10");
	});

	it.each([
		"0",
		"-1",
		"",
	])("disables the start button when Repeat Count is %j", async (value) => {
		const user = userEvent.setup();
		render(<App />);

		const input = screen.getByLabelText(/repeat count/i);
		await user.clear(input);
		if (value !== "") {
			await user.type(input, value);
		}

		expect(screen.getByRole("button", { name: /テスト撮影/ })).toBeDisabled();
	});

	it("passes the entered Repeat Count to RunTestSession", async () => {
		const user = userEvent.setup();
		vi.mocked(RunTestSession).mockClear();
		render(<App />);

		const input = screen.getByLabelText(/repeat count/i);
		await user.clear(input);
		await user.type(input, "7");
		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(RunTestSession).toHaveBeenCalledWith(7);
	});
});
