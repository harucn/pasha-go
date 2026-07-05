import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	ChooseOutputDirectory,
	DefaultOutputFileName,
	RunTestSession,
} from "../wailsjs/go/main/App";
import {
	WindowGetPosition,
	WindowGetSize,
	WindowSetPosition,
	WindowSetSize,
} from "../wailsjs/runtime/runtime";

vi.mock("../wailsjs/go/main/App", () => ({
	RunTestSession: vi.fn(() => Promise.resolve()),
	DefaultOutputFileName: vi.fn(() => Promise.resolve("pasha-2026-06-28_15-30")),
	ChooseOutputDirectory: vi.fn(() => Promise.resolve("")),
}));

vi.mock("../wailsjs/runtime/runtime", () => ({
	WindowSetSize: vi.fn(),
	WindowSetPosition: vi.fn(),
	WindowGetSize: vi.fn(() => Promise.resolve({ w: 800, h: 600 })),
	WindowGetPosition: vi.fn(() => Promise.resolve({ x: 100, y: 100 })),
}));

import App from "./App";

type RegionGeometry = { x: number; y: number; width: number; height: number };

async function selectRegion(
	user: ReturnType<typeof userEvent.setup>,
	region: RegionGeometry = { x: 10, y: 20, width: 100, height: 50 },
) {
	await user.click(screen.getByRole("button", { name: /範囲選択/ }));
	await screen.findByRole("dialog");
	vi.mocked(WindowGetPosition).mockResolvedValueOnce({
		x: region.x,
		y: region.y,
	});
	vi.mocked(WindowGetSize).mockResolvedValueOnce({
		w: region.width,
		h: region.height,
	});
	await user.click(screen.getByRole("button", { name: /確定/ }));
	await waitFor(() => {
		expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
	});
}

beforeEach(() => {
	vi.mocked(RunTestSession).mockClear();
	vi.mocked(ChooseOutputDirectory).mockClear();
	vi.mocked(DefaultOutputFileName)
		.mockClear()
		.mockResolvedValue("pasha-2026-06-28_15-30");
	vi.mocked(WindowSetSize).mockClear();
	vi.mocked(WindowSetPosition).mockClear();
	vi.mocked(WindowGetSize).mockClear().mockResolvedValue({ w: 800, h: 600 });
	vi.mocked(WindowGetPosition)
		.mockClear()
		.mockResolvedValue({ x: 100, y: 100 });
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

		await selectRegion(user);

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

	it("renders a 範囲選択 (region select) button", () => {
		render(<App />);
		expect(
			screen.getByRole("button", { name: /範囲選択/ }),
		).toBeInTheDocument();
	});

	it("shrinks the window and shows the region frame when 範囲選択 is clicked", async () => {
		const user = userEvent.setup();
		render(<App />);

		expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

		await user.click(screen.getByRole("button", { name: /範囲選択/ }));

		expect(await screen.findByRole("dialog")).toBeInTheDocument();
		await waitFor(() => {
			expect(WindowSetSize).toHaveBeenCalledWith(500, 400);
		});
	});

	it("records the window geometry as the region when 確定 is clicked and restores the window", async () => {
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /範囲選択/ }));
		await screen.findByRole("dialog");

		vi.mocked(WindowGetPosition).mockResolvedValueOnce({ x: 250, y: 180 });
		vi.mocked(WindowGetSize).mockResolvedValueOnce({ w: 640, h: 320 });

		await user.click(screen.getByRole("button", { name: /確定/ }));

		await waitFor(() => {
			expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
		});
		expect(WindowSetPosition).toHaveBeenLastCalledWith(100, 100);
		expect(WindowSetSize).toHaveBeenLastCalledWith(800, 600);
		expect(
			await screen.findByText(/範囲指定済み.*250.*180.*640.*320/),
		).toBeInTheDocument();
	});

	it("cancels region selection when Escape is pressed", async () => {
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /範囲選択/ }));
		await screen.findByRole("dialog");

		await user.keyboard("{Escape}");

		await waitFor(() => {
			expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
		});
		expect(WindowSetPosition).toHaveBeenLastCalledWith(100, 100);
		expect(WindowSetSize).toHaveBeenLastCalledWith(800, 600);
		expect(screen.queryByText(/範囲指定済み/)).not.toBeInTheDocument();
	});

	it("cancels region selection when キャンセル button is clicked", async () => {
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /範囲選択/ }));
		await screen.findByRole("dialog");

		await user.click(screen.getByRole("button", { name: /キャンセル/ }));

		await waitFor(() => {
			expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
		});
		expect(screen.queryByText(/範囲指定済み/)).not.toBeInTheDocument();
	});

	it("keeps the start button disabled until a capture region is selected", async () => {
		vi.mocked(ChooseOutputDirectory).mockResolvedValueOnce("/tmp/out");
		const user = userEvent.setup();
		render(<App />);

		await waitFor(() => {
			const input = screen.getByLabelText(/file name/i) as HTMLInputElement;
			expect(input.value).toBe("pasha-2026-06-28_15-30");
		});

		await user.click(screen.getByRole("button", { name: /folder|フォルダ/i }));
		await screen.findByText("/tmp/out");

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

		await selectRegion(user);

		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(RunTestSession).toHaveBeenCalledWith({
			repeatCount: 7,
			stepIntervalSeconds: 2.5,
			outputDir: "/tmp/out",
			outputFileName: "custom-name",
			captureRegion: { x: 10, y: 20, width: 100, height: 50 },
		});
	});

	it("overwrites the previously selected region on re-selection", async () => {
		vi.mocked(ChooseOutputDirectory).mockResolvedValueOnce("/tmp/out");
		const user = userEvent.setup();
		render(<App />);
		await waitFor(() => {
			const input = screen.getByLabelText(/file name/i) as HTMLInputElement;
			expect(input.value).toBe("pasha-2026-06-28_15-30");
		});
		await user.click(screen.getByRole("button", { name: /folder|フォルダ/i }));
		await screen.findByText("/tmp/out");

		await selectRegion(user);
		await selectRegion(user, { x: 500, y: 600, width: 200, height: 300 });

		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(RunTestSession).toHaveBeenLastCalledWith(
			expect.objectContaining({
				captureRegion: { x: 500, y: 600, width: 200, height: 300 },
			}),
		);
	});

	it("shows a region-selected indicator after a region has been picked", async () => {
		const user = userEvent.setup();
		render(<App />);

		expect(screen.queryByText(/範囲指定済み/)).not.toBeInTheDocument();

		await selectRegion(user);

		expect(await screen.findByText(/範囲指定済み/)).toBeInTheDocument();
	});
});
