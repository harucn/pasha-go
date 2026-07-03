import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	ChooseOutputDirectory,
	DefaultOutputFileName,
	RunTestSession,
} from "../wailsjs/go/main/App";

vi.mock("../wailsjs/go/main/App", () => ({
	RunTestSession: vi.fn(() => Promise.resolve()),
	DefaultOutputFileName: vi.fn(() => Promise.resolve("pasha-2026-06-28_15-30")),
	ChooseOutputDirectory: vi.fn(() => Promise.resolve("")),
}));

import App from "./App";

beforeEach(() => {
	vi.mocked(RunTestSession).mockClear();
	vi.mocked(ChooseOutputDirectory).mockClear();
	vi.mocked(DefaultOutputFileName)
		.mockClear()
		.mockResolvedValue("pasha-2026-06-28_15-30");
});

describe("App", () => {
	it("renders the initial prompt", () => {
		render(<App />);
		expect(screen.getByText(/press the button/i)).toBeInTheDocument();
	});

	it("shows completion message referring to the chosen output path", async () => {
		vi.mocked(ChooseOutputDirectory).mockResolvedValueOnce("/tmp/out");
		const user = userEvent.setup();
		render(<App />);
		await waitFor(() => {
			const input = screen.getByLabelText(/file name/i) as HTMLInputElement;
			expect(input.value).toBe("pasha-2026-06-28_15-30");
		});

		await user.click(screen.getByRole("button", { name: /folder|フォルダ/i }));
		await screen.findByText("/tmp/out");

		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(
			await screen.findByText(/\/tmp\/out\/pasha-2026-06-28_15-30\.pdf/),
		).toBeInTheDocument();
	});

	it("renders a Repeat Count input with default value 10", () => {
		render(<App />);
		const input = screen.getByLabelText(/repeat count/i) as HTMLInputElement;
		expect(input).toBeInTheDocument();
		expect(input.type).toBe("number");
		expect(input.value).toBe("10");
	});

	it("renders a Step Interval input with default value 1.0 seconds", () => {
		render(<App />);
		const input = screen.getByLabelText(/step interval/i) as HTMLInputElement;
		expect(input).toBeInTheDocument();
		expect(input.type).toBe("number");
		expect(input.value).toBe("1.0");
	});

	it("populates the Output File Name input from DefaultOutputFileName", async () => {
		render(<App />);
		const input = (await screen.findByLabelText(
			/file name/i,
		)) as HTMLInputElement;
		await waitFor(() => {
			expect(input.value).toBe("pasha-2026-06-28_15-30");
		});
	});

	it("chooses an output directory and displays the selected path", async () => {
		vi.mocked(ChooseOutputDirectory).mockResolvedValueOnce(
			"/Users/foo/Documents",
		);
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /folder|フォルダ/i }));

		expect(await screen.findByText("/Users/foo/Documents")).toBeInTheDocument();
	});

	it("keeps the start button disabled until a folder has been chosen", async () => {
		render(<App />);
		await waitFor(() => {
			const input = screen.getByLabelText(/file name/i) as HTMLInputElement;
			expect(input.value).toBe("pasha-2026-06-28_15-30");
		});

		expect(screen.getByRole("button", { name: /テスト撮影/ })).toBeDisabled();
	});

	it("disables the start button when Output File Name is empty", async () => {
		vi.mocked(ChooseOutputDirectory).mockResolvedValueOnce("/tmp/out");
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /folder|フォルダ/i }));
		await screen.findByText("/tmp/out");

		const fileNameInput = screen.getByLabelText(/file name/i);
		await user.clear(fileNameInput);

		expect(screen.getByRole("button", { name: /テスト撮影/ })).toBeDisabled();
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

	it.each([
		"0",
		"-0.5",
		"",
	])("disables the start button when Step Interval is %j", async (value) => {
		const user = userEvent.setup();
		render(<App />);

		const input = screen.getByLabelText(/step interval/i);
		await user.clear(input);
		if (value !== "") {
			await user.type(input, value);
		}

		expect(screen.getByRole("button", { name: /テスト撮影/ })).toBeDisabled();
	});

	it("passes all inputs as a params object to RunTestSession", async () => {
		vi.mocked(ChooseOutputDirectory).mockResolvedValueOnce("/tmp/out");
		const user = userEvent.setup();
		render(<App />);
		await waitFor(() => {
			const input = screen.getByLabelText(/file name/i) as HTMLInputElement;
			expect(input.value).toBe("pasha-2026-06-28_15-30");
		});

		await user.click(screen.getByRole("button", { name: /folder|フォルダ/i }));
		await screen.findByText("/tmp/out");

		const repeatInput = screen.getByLabelText(/repeat count/i);
		await user.clear(repeatInput);
		await user.type(repeatInput, "7");

		const intervalInput = screen.getByLabelText(/step interval/i);
		await user.clear(intervalInput);
		await user.type(intervalInput, "2.5");

		const fileNameInput = screen.getByLabelText(/file name/i);
		await user.clear(fileNameInput);
		await user.type(fileNameInput, "custom-name");

		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(RunTestSession).toHaveBeenCalledWith({
			repeatCount: 7,
			stepIntervalSeconds: 2.5,
			outputDir: "/tmp/out",
			outputFileName: "custom-name",
		});
	});
});
