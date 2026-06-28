import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

vi.mock("../wailsjs/go/main/App", () => ({
	Greet: vi.fn((name: string) =>
		Promise.resolve(`Hello ${name}, It's show time!`),
	),
}));

import App from "./App";

describe("App", () => {
	it("renders the initial prompt", () => {
		render(<App />);
		expect(screen.getByText(/please enter your name/i)).toBeInTheDocument();
	});

	it("shows the greeting after clicking Greet", async () => {
		const user = userEvent.setup();
		render(<App />);

		await user.type(screen.getByRole("textbox"), "World");
		await user.click(screen.getByRole("button", { name: /greet/i }));

		expect(
			await screen.findByText("Hello World, It's show time!"),
		).toBeInTheDocument();
	});
});
