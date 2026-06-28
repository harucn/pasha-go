import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

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
});
