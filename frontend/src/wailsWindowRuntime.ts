import {
	WindowGetPosition,
	WindowGetSize,
	WindowSetMaxSize,
	WindowSetMinSize,
	WindowSetPosition,
	WindowSetSize,
} from "../wailsjs/runtime/runtime";
import type { Geometry, Point, Size, WindowRuntime } from "./selectionWindow";

// Wails pins a window by setting its min and max size to the same value, and
// reads a zero max size as "no maximum". Nothing outside this adapter should
// have to know that.
const NO_MAXIMUM: Size = { w: 0, h: 0 };

// Small enough that the user can shrink the Selection Window to any Capture
// Region worth capturing.
const MINIMUM_WHILE_UNLOCKED: Size = { w: 200, h: 150 };

export const wailsWindowRuntime: WindowRuntime = {
	async geometry(): Promise<Geometry> {
		const [size, pos] = await Promise.all([
			WindowGetSize(),
			WindowGetPosition(),
		]);
		return { size, pos };
	},

	resize(size: Size) {
		WindowSetSize(size.w, size.h);
	},

	moveTo(pos: Point) {
		WindowSetPosition(pos.x, pos.y);
	},

	unlockSize() {
		WindowSetMinSize(MINIMUM_WHILE_UNLOCKED.w, MINIMUM_WHILE_UNLOCKED.h);
		WindowSetMaxSize(NO_MAXIMUM.w, NO_MAXIMUM.h);
	},

	lockSizeTo(size: Size) {
		WindowSetMinSize(size.w, size.h);
		WindowSetMaxSize(size.w, size.h);
	},
};
