import { describe, expect, it, vi } from "vitest";
import type { main } from "../wailsjs/go/models";
import {
	createSelectionWindow,
	frameSize,
	type Geometry,
	type Point,
	type Size,
	type WindowRuntime,
} from "./selectionWindow";

const BAR: Geometry = { size: { w: 800, h: 600 }, pos: { x: 100, y: 100 } };

/**
 * Models the window as state rather than as a call log: `locked` is what the
 * user can actually do to the window, which is what the invariant is about.
 */
type FakeRuntime = WindowRuntime & {
	readonly size: Size;
	readonly pos: Point;
	readonly locked: Size | null;
};

function fakeRuntime(initial: Geometry = BAR): FakeRuntime {
	let size: Size = { ...initial.size };
	let pos: Point = { ...initial.pos };
	let locked: Size | null = { ...initial.size };

	return {
		get size() {
			return size;
		},
		get pos() {
			return pos;
		},
		get locked() {
			return locked;
		},

		async geometry(): Promise<Geometry> {
			return { size: { ...size }, pos: { ...pos } };
		},
		resize(next: Size) {
			size = { ...next };
		},
		moveTo(next: Point) {
			pos = { ...next };
		},
		unlockSize() {
			locked = null;
		},
		lockSizeTo(next: Size) {
			locked = { ...next };
		},
	};
}

const aSelection = {
	region: { x: 250, y: 180, width: 640, height: 320 },
	clickPoint: { x: 300, y: 250 },
} as main.RegionSelection;

describe("open", () => {
	it("unlocks the size and resizes to the frame", async () => {
		const rt = fakeRuntime();
		const sw = createSelectionWindow(rt, vi.fn());

		await sw.open();

		expect(rt.locked).toBeNull();
		expect(rt.size).toEqual(frameSize);
	});

	it("refuses to open twice, keeping the stashed bar geometry", async () => {
		const rt = fakeRuntime();
		const sw = createSelectionWindow(rt, vi.fn());

		await sw.open();
		await expect(sw.open()).rejects.toThrow(/already open/);

		// The bar's geometry survived, so cancelling still restores the bar and
		// not the frame it was showing.
		sw.cancel();
		expect(rt.size).toEqual(BAR.size);
		expect(rt.locked).toEqual(BAR.size);
	});
});

describe("cancel", () => {
	it("restores the bar's geometry and re-locks its size", async () => {
		const rt = fakeRuntime();
		const sw = createSelectionWindow(rt, vi.fn());

		await sw.open();
		rt.resize({ w: 320, h: 240 }); // the user dragged the frame smaller
		sw.cancel();

		expect(rt.size).toEqual(BAR.size);
		expect(rt.pos).toEqual(BAR.pos);
		expect(rt.locked).toEqual(BAR.size);
	});

	it("is a no-op when the Selection Window was never opened", () => {
		const rt = fakeRuntime();
		const sw = createSelectionWindow(rt, vi.fn());

		sw.cancel();

		expect(rt.size).toEqual(BAR.size);
		expect(rt.locked).toEqual(BAR.size);
	});

	it("can be followed by another open", async () => {
		const rt = fakeRuntime();
		const sw = createSelectionWindow(rt, vi.fn());

		await sw.open();
		sw.cancel();
		await sw.open();

		expect(rt.locked).toBeNull();
		expect(rt.size).toEqual(frameSize);
	});
});

describe("confirm", () => {
	it("reports the selection and restores the bar", async () => {
		const rt = fakeRuntime();
		const getSelection = vi.fn(async () => aSelection);
		const sw = createSelectionWindow(rt, getSelection);

		await sw.open();
		const got = await sw.confirm({ x: 280, y: 240 });

		expect(got).toBe(aSelection);
		expect(getSelection).toHaveBeenCalledWith(280, 240);
		expect(rt.size).toEqual(BAR.size);
		expect(rt.pos).toEqual(BAR.pos);
		expect(rt.locked).toEqual(BAR.size);
	});

	// The failure mode the module exists to make impossible: a throw partway
	// through confirming must not leave the bar resizable.
	it("restores the bar even when reading the selection fails", async () => {
		const rt = fakeRuntime();
		const boom = new Error("no main window available");
		const sw = createSelectionWindow(
			rt,
			vi.fn(() => Promise.reject(boom)),
		);

		await sw.open();
		await expect(sw.confirm({ x: 10, y: 10 })).rejects.toThrow(boom);

		expect(rt.size).toEqual(BAR.size);
		expect(rt.pos).toEqual(BAR.pos);
		expect(rt.locked).toEqual(BAR.size);
	});

	it("can be followed by another open", async () => {
		const rt = fakeRuntime();
		const sw = createSelectionWindow(
			rt,
			vi.fn(async () => aSelection),
		);

		await sw.open();
		await sw.confirm({ x: 1, y: 1 });
		await sw.open();

		expect(rt.locked).toBeNull();
	});
});
