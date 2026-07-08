import {
	fireEvent,
	render,
	screen,
	waitFor,
	within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	ChooseOutputDirectory,
	DefaultOutputFileName,
	GetSelectedRegion,
	RunTestSession,
	StopSession,
} from "../wailsjs/go/main/App";
import {
	EventsOn,
	WindowGetPosition,
	WindowGetSize,
	WindowSetMaxSize,
	WindowSetMinSize,
	WindowSetPosition,
	WindowSetSize,
} from "../wailsjs/runtime/runtime";

vi.mock("../wailsjs/go/main/App", () => ({
	RunTestSession: vi.fn(() => Promise.resolve()),
	StopSession: vi.fn(() => Promise.resolve()),
	DefaultOutputFileName: vi.fn(() => Promise.resolve("pasha-2026-06-28_15-30")),
	ChooseOutputDirectory: vi.fn(() => Promise.resolve("")),
	GetSelectedRegion: vi.fn(() =>
		Promise.resolve({ x: 10, y: 20, width: 100, height: 50 }),
	),
}));

vi.mock("../wailsjs/runtime/runtime", () => ({
	WindowSetSize: vi.fn(),
	WindowSetPosition: vi.fn(),
	WindowSetMinSize: vi.fn(),
	WindowSetMaxSize: vi.fn(),
	WindowGetSize: vi.fn(() => Promise.resolve({ w: 800, h: 600 })),
	WindowGetPosition: vi.fn(() => Promise.resolve({ x: 100, y: 100 })),
	EventsOn: vi.fn(() => () => {}),
}));

import App from "./App";

type RegionGeometry = { x: number; y: number; width: number; height: number };

// selectRegion opens the region-selection dialog, mocks GetSelectedRegion,
// optionally drags the click-point marker by the given delta (from its
// initial center at (REGION_FRAME_WIDTH/2, REGION_FRAME_HEIGHT/2) = (250, 200)),
// and confirms. After this the App state has BOTH region and clickPoint set
// atomically (clickPoint = region.min + finalMarkerPos).
async function selectRegion(
	user: ReturnType<typeof userEvent.setup>,
	region: RegionGeometry = { x: 10, y: 20, width: 100, height: 50 },
	markerDelta: { dx: number; dy: number } = { dx: 0, dy: 0 },
) {
	await user.click(screen.getByRole("button", { name: /範囲選択/ }));
	await screen.findByRole("dialog", { name: /Capture Region selection/ });
	vi.mocked(GetSelectedRegion).mockResolvedValueOnce(region);
	const dialog = screen.getByRole("dialog", {
		name: /Capture Region selection/,
	});
	if (markerDelta.dx !== 0 || markerDelta.dy !== 0) {
		const marker = within(dialog).getByLabelText(
			/クリック位置マーカー/,
		) as HTMLElement;
		fireEvent.pointerDown(marker, { clientX: 0, clientY: 0, pointerId: 1 });
		fireEvent.pointerMove(marker, {
			clientX: markerDelta.dx,
			clientY: markerDelta.dy,
			pointerId: 1,
		});
		fireEvent.pointerUp(marker, {
			clientX: markerDelta.dx,
			clientY: markerDelta.dy,
			pointerId: 1,
		});
	}
	await user.click(within(dialog).getByRole("button", { name: /確定/ }));
	await waitFor(() => {
		expect(
			screen.queryByRole("dialog", { name: /Capture Region selection/ }),
		).not.toBeInTheDocument();
	});
}

beforeEach(() => {
	vi.mocked(RunTestSession).mockClear().mockResolvedValue(undefined);
	vi.mocked(StopSession).mockClear().mockResolvedValue(undefined);
	vi.mocked(ChooseOutputDirectory).mockClear();
	vi.mocked(DefaultOutputFileName)
		.mockClear()
		.mockResolvedValue("pasha-2026-06-28_15-30");
	vi.mocked(GetSelectedRegion)
		.mockClear()
		.mockResolvedValue({ x: 10, y: 20, width: 100, height: 50 });
	vi.mocked(WindowSetSize).mockClear();
	vi.mocked(WindowSetPosition).mockClear();
	vi.mocked(WindowSetMinSize).mockClear();
	vi.mocked(WindowSetMaxSize).mockClear();
	vi.mocked(WindowGetSize).mockClear().mockResolvedValue({ w: 800, h: 600 });
	vi.mocked(WindowGetPosition)
		.mockClear()
		.mockResolvedValue({ x: 100, y: 100 });
	vi.mocked(EventsOn)
		.mockClear()
		.mockReturnValue(() => {});
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

	it("records the region from GetSelectedRegion when 確定 is clicked and restores the window", async () => {
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /範囲選択/ }));
		await screen.findByRole("dialog");

		vi.mocked(GetSelectedRegion).mockResolvedValueOnce({
			x: 250,
			y: 180,
			width: 640,
			height: 320,
		});

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

	it("shows an advance click point marker inside the region-selection dialog", async () => {
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /範囲選択/ }));
		const dialog = await screen.findByRole("dialog", {
			name: /Capture Region selection/,
		});

		expect(
			within(dialog).getByLabelText(/クリック位置マーカー/),
		).toBeInTheDocument();
	});

	it("moves the marker on drag by the pointer delta", async () => {
		const user = userEvent.setup();
		render(<App />);

		await user.click(screen.getByRole("button", { name: /範囲選択/ }));
		const dialog = await screen.findByRole("dialog", {
			name: /Capture Region selection/,
		});
		const marker = within(dialog).getByLabelText(
			/クリック位置マーカー/,
		) as HTMLElement;

		// Region frame is 500x400. Initial marker position is center (250, 200).
		fireEvent.pointerDown(marker, { clientX: 100, clientY: 80, pointerId: 1 });
		fireEvent.pointerMove(marker, { clientX: 150, clientY: 130, pointerId: 1 });
		fireEvent.pointerUp(marker, { clientX: 150, clientY: 130, pointerId: 1 });

		expect(marker.style.left).toBe("300px");
		expect(marker.style.top).toBe("250px");
	});

	it("saves both region and click point when confirming the region-selection dialog", async () => {
		vi.mocked(ChooseOutputDirectory).mockResolvedValueOnce("/tmp/out");
		const user = userEvent.setup();
		render(<App />);

		await waitFor(() => {
			const input = screen.getByLabelText(/file name/i) as HTMLInputElement;
			expect(input.value).toBe("pasha-2026-06-28_15-30");
		});
		await user.click(screen.getByRole("button", { name: /folder|フォルダ/i }));
		await screen.findByText("/tmp/out");

		await user.click(screen.getByRole("button", { name: /範囲選択/ }));
		const dialog = await screen.findByRole("dialog", {
			name: /Capture Region selection/,
		});
		const marker = within(dialog).getByLabelText(
			/クリック位置マーカー/,
		) as HTMLElement;

		// Default marker (250, 200) → drag to (300, 250) via delta +50, +50.
		fireEvent.pointerDown(marker, { clientX: 100, clientY: 80, pointerId: 1 });
		fireEvent.pointerMove(marker, { clientX: 150, clientY: 130, pointerId: 1 });
		fireEvent.pointerUp(marker, { clientX: 150, clientY: 130, pointerId: 1 });

		// GetSelectedRegion mock returns { x: 10, y: 20, width: 100, height: 50 }.
		await user.click(within(dialog).getByRole("button", { name: /確定/ }));
		await waitFor(() => {
			expect(
				screen.queryByRole("dialog", { name: /Capture Region selection/ }),
			).not.toBeInTheDocument();
		});

		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(RunTestSession).toHaveBeenCalledWith(
			expect.objectContaining({
				captureRegion: { x: 10, y: 20, width: 100, height: 50 },
				advanceClickPoint: { x: 310, y: 270 },
			}),
		);
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

	it("keeps the start button disabled until an advance click point is selected", async () => {
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

		await selectRegion(user);

		// After region confirm, both region and clickPoint are set atomically,
		// so the button becomes enabled.
		expect(
			screen.getByRole("button", { name: /テスト撮影/ }),
		).not.toBeDisabled();
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

		// Drag marker +30, +40 → marker at (280, 240). Region at (10, 20).
		// Expected clickPoint = (10 + 280, 20 + 240) = (290, 260).
		await selectRegion(
			user,
			{ x: 10, y: 20, width: 100, height: 50 },
			{ dx: 30, dy: 40 },
		);

		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(RunTestSession).toHaveBeenCalledWith({
			repeatCount: 7,
			stepIntervalSeconds: 2.5,
			outputDir: "/tmp/out",
			outputFileName: "custom-name",
			captureRegion: { x: 10, y: 20, width: 100, height: 50 },
			advanceClickPoint: { x: 290, y: 260 },
		});
	});

	it("overwrites the previously selected region and click point on re-selection", async () => {
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
		// Second confirm — marker resets to center each time the dialog opens.
		// Region (500, 600), marker default (250, 200) → clickPoint (750, 800).
		await selectRegion(user, { x: 500, y: 600, width: 200, height: 300 });

		await user.click(screen.getByRole("button", { name: /テスト撮影/ }));

		expect(RunTestSession).toHaveBeenLastCalledWith(
			expect.objectContaining({
				captureRegion: { x: 500, y: 600, width: 200, height: 300 },
				advanceClickPoint: { x: 750, y: 800 },
			}),
		);
	});

	it("renders the main controls inside a toolbar landmark (floating bar)", () => {
		render(<App />);
		const bar = screen.getByRole("toolbar", { name: /pasha controls/i });
		expect(bar).toBeInTheDocument();
		expect(within(bar).getByLabelText(/repeat count/i)).toBeInTheDocument();
		expect(
			within(bar).getByRole("button", { name: /範囲選択/ }),
		).toBeInTheDocument();
		expect(
			within(bar).getByRole("button", { name: /テスト撮影/ }),
		).toBeInTheDocument();
	});

	it("does not render the legacy logo image (bar has no room for it)", () => {
		render(<App />);
		expect(screen.queryByAltText(/logo/i)).not.toBeInTheDocument();
	});

	it("shows live progress from session:progress events on the bar", async () => {
		let progressHandler: ((data: unknown) => void) | undefined;
		vi.mocked(EventsOn).mockImplementation((event, handler) => {
			if (event === "session:progress") {
				progressHandler = handler as (data: unknown) => void;
			}
			return () => {};
		});

		render(<App />);

		expect(progressHandler).toBeDefined();

		progressHandler?.({ current: 3, total: 10 });
		expect(await screen.findByText(/3\s*\/\s*10/)).toBeInTheDocument();

		progressHandler?.({ current: 10, total: 10 });
		expect(await screen.findByText(/10\s*\/\s*10/)).toBeInTheDocument();
	});

	it("swaps テスト撮影 for a 停止 button while a session runs and calls StopSession on click", async () => {
		vi.mocked(ChooseOutputDirectory).mockResolvedValueOnce("/tmp/out");
		// Keep the session pending so `running` stays true and the 停止 button
		// remains on screen.
		let resolveRun: (() => void) | undefined;
		vi.mocked(RunTestSession).mockImplementationOnce(
			() =>
				new Promise<void>((res) => {
					resolveRun = () => res();
				}),
		);
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

		const stopButton = await screen.findByRole("button", { name: /停止/ });
		expect(
			screen.queryByRole("button", { name: /テスト撮影/ }),
		).not.toBeInTheDocument();

		await user.click(stopButton);
		expect(StopSession).toHaveBeenCalled();

		resolveRun?.();
	});

	it("shows a 撮影終了 state on the bar when a session:completed event arrives", async () => {
		let completedHandler: (() => void) | undefined;
		vi.mocked(EventsOn).mockImplementation((event, handler) => {
			if (event === "session:completed") {
				completedHandler = handler as () => void;
			}
			return () => {};
		});

		render(<App />);

		expect(completedHandler).toBeDefined();
		completedHandler?.();

		expect(await screen.findByText(/撮影終了/)).toBeInTheDocument();
	});

	it("shows a red inline error on the bar when a session:error event arrives", async () => {
		let errorHandler: ((data: unknown) => void) | undefined;
		vi.mocked(EventsOn).mockImplementation((event, handler) => {
			if (event === "session:error") {
				errorHandler = handler as (data: unknown) => void;
			}
			return () => {};
		});

		render(<App />);

		expect(errorHandler).toBeDefined();
		errorHandler?.({ message: "スクリーンキャプチャに失敗しました。" });

		const alert = await screen.findByRole("alert");
		expect(alert).toHaveTextContent("スクリーンキャプチャに失敗しました。");
	});

	it("dismisses the error when the close button is clicked", async () => {
		let errorHandler: ((data: unknown) => void) | undefined;
		vi.mocked(EventsOn).mockImplementation((event, handler) => {
			if (event === "session:error") {
				errorHandler = handler as (data: unknown) => void;
			}
			return () => {};
		});
		const user = userEvent.setup();
		render(<App />);
		errorHandler?.({ message: "PDF の書き込みに失敗しました。" });

		await user.click(await screen.findByRole("button", { name: /閉じる/ }));

		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("shows a region-selected indicator after a region has been picked", async () => {
		const user = userEvent.setup();
		render(<App />);

		expect(screen.queryByText(/範囲指定済み/)).not.toBeInTheDocument();

		await selectRegion(user);

		expect(await screen.findByText(/範囲指定済み/)).toBeInTheDocument();
	});
});
