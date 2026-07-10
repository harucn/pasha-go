import type { main } from "../wailsjs/go/models";

/**
 * The Selection Window: the floating bar, temporarily reshaped into a resizable
 * frame the user drags around the Capture Region.
 *
 * Its one invariant is that unlocking and restoring are a pair. `open` stashes
 * the bar's geometry and releases its size lock; the only two ways back out —
 * `confirm` and `cancel` — both put it back. `confirm` restores even when the
 * selection itself fails, so no failure mode leaves the bar resizable.
 */

export type Size = { w: number; h: number };
export type Point = { x: number; y: number };
export type Geometry = { size: Size; pos: Point };

/** The size the Selection Window opens at, and so the size of its frame. */
export const frameSize: Size = { w: 500, h: 400 };

/**
 * The window operations the Selection Window needs, named for what they mean
 * rather than for the Wails calls behind them.
 */
export interface WindowRuntime {
	geometry(): Promise<Geometry>;
	resize(size: Size): void;
	moveTo(pos: Point): void;
	/** Let the user resize freely. */
	unlockSize(): void;
	/** Pin the window to exactly this size. */
	lockSizeTo(size: Size): void;
}

/** Reads the Capture Region and Advance Click Point from Go, in Screen Space. */
export type GetSelection = (
	offsetX: number,
	offsetY: number,
) => Promise<main.RegionSelection>;

export interface SelectionWindow {
	open(): Promise<void>;
	confirm(markerOffset: Point): Promise<main.RegionSelection>;
	cancel(): void;
}

export function createSelectionWindow(
	runtime: WindowRuntime,
	getSelection: GetSelection,
): SelectionWindow {
	// Non-null exactly while the Selection Window is open.
	let stashed: Geometry | null = null;

	function restore(): void {
		if (!stashed) return;
		runtime.moveTo(stashed.pos);
		runtime.resize(stashed.size);
		// Re-pin the bar so the user cannot resize it once it is a bar again.
		runtime.lockSizeTo(stashed.size);
		stashed = null;
	}

	return {
		async open() {
			// Opening twice would stash the frame's own geometry over the bar's,
			// stranding the bar at the frame's size for the rest of the run.
			if (stashed) {
				throw new Error("selectionWindow: already open");
			}
			stashed = await runtime.geometry();
			runtime.unlockSize();
			runtime.resize(frameSize);
		},

		async confirm(markerOffset: Point) {
			try {
				return await getSelection(markerOffset.x, markerOffset.y);
			} finally {
				restore();
			}
		},

		cancel() {
			restore();
		},
	};
}
