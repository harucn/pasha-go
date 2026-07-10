import "./App.css";
import {
	type PointerEvent as ReactPointerEvent,
	useCallback,
	useEffect,
	useRef,
	useState,
} from "react";
import {
	ChooseOutputDirectory,
	DefaultOutputFileName,
	GetSelectedRegion,
	RunTestSession,
	StopSession,
} from "../wailsjs/go/main/App";
import { main } from "../wailsjs/go/models";
import {
	EventsOn,
	WindowGetPosition,
	WindowGetSize,
	WindowSetMaxSize,
	WindowSetMinSize,
	WindowSetPosition,
	WindowSetSize,
} from "../wailsjs/runtime/runtime";

type CaptureRegion = {
	x: number;
	y: number;
	width: number;
	height: number;
};

const REGION_FRAME_WIDTH = 500;
const REGION_FRAME_HEIGHT = 400;

// Under the field's `direction: rtl`, bidi resolution sweeps a POSIX path's
// leading "/" to the far right. A LEFT-TO-RIGHT MARK anchors it. Display only.
const LRM = "\u200e";

function displayPath(path: string): string {
	return LRM + path;
}

function App() {
	// Empty until a Capture Session reports something. The status line keeps
	// its height so the toolbar below does not shift when a message appears.
	const [status, setStatus] = useState("");
	const [running, setRunning] = useState(false);
	const [errorMessage, setErrorMessage] = useState<string | null>(null);
	const [repeatCount, setRepeatCount] = useState("10");
	const [stepInterval, setStepInterval] = useState("1.0");
	const [outputFileName, setOutputFileName] = useState("");
	const [outputDir, setOutputDir] = useState("");
	const [selectingRegion, setSelectingRegion] = useState(false);
	const [region, setRegion] = useState<CaptureRegion | null>(null);
	const [clickPoint, setClickPoint] = useState<{ x: number; y: number } | null>(
		null,
	);
	const [markerPos, setMarkerPos] = useState({
		x: REGION_FRAME_WIDTH / 2,
		y: REGION_FRAME_HEIGHT / 2,
	});
	const dragStartRef = useRef<{
		startX: number;
		startY: number;
		markerX: number;
		markerY: number;
	} | null>(null);
	const originalWindowRef = useRef<{
		size: { w: number; h: number };
		pos: { x: number; y: number };
	} | null>(null);

	const handleMarkerPointerDown = useCallback(
		(e: ReactPointerEvent<HTMLButtonElement>) => {
			dragStartRef.current = {
				startX: e.clientX,
				startY: e.clientY,
				markerX: markerPos.x,
				markerY: markerPos.y,
			};
			e.currentTarget.setPointerCapture?.(e.pointerId);
		},
		[markerPos.x, markerPos.y],
	);

	const handleMarkerPointerMove = useCallback(
		(e: ReactPointerEvent<HTMLButtonElement>) => {
			const start = dragStartRef.current;
			if (!start) return;
			setMarkerPos({
				x: start.markerX + (e.clientX - start.startX),
				y: start.markerY + (e.clientY - start.startY),
			});
		},
		[],
	);

	const handleMarkerPointerUp = useCallback(
		(e: ReactPointerEvent<HTMLButtonElement>) => {
			dragStartRef.current = null;
			if (e.currentTarget.hasPointerCapture?.(e.pointerId)) {
				e.currentTarget.releasePointerCapture(e.pointerId);
			}
		},
		[],
	);

	const restoreWindow = useCallback(() => {
		const orig = originalWindowRef.current;
		if (orig) {
			WindowSetPosition(orig.pos.x, orig.pos.y);
			WindowSetSize(orig.size.w, orig.size.h);
			// Re-lock the bar to its original (fixed) dimensions so the user
			// cannot resize it after returning from Capture Region selection.
			WindowSetMinSize(orig.size.w, orig.size.h);
			WindowSetMaxSize(orig.size.w, orig.size.h);
		}
	}, []);

	const cancelRegionSelection = useCallback(() => {
		restoreWindow();
		setSelectingRegion(false);
	}, [restoreWindow]);

	useEffect(() => {
		DefaultOutputFileName().then(setOutputFileName);
	}, []);

	// Subscribe to Capture Session progress ticks emitted from Go
	// (app.go emitProgress). The bar is too narrow for a progress bar,
	// so we render "N / M steps" as text (issue #08).
	useEffect(() => {
		const off = EventsOn("session:progress", (data: unknown) => {
			const p = data as { current?: number; total?: number };
			if (typeof p?.current === "number" && typeof p?.total === "number") {
				setStatus(`${p.current} / ${p.total} steps`);
			}
		});
		return off;
	}, []);

	// Subscribe to the Capture Session completion signal (issue #09). Fired
	// whether the session ran to its Repeat Count or was stopped early; either
	// way the bar transitions to the finished state.
	//
	// The status line is written by runTestSession instead, from the Output
	// Document path RunTestSession returns. Writing it here too would race:
	// the arrival order of a Wails event and a resolved binding promise is
	// not guaranteed.
	useEffect(() => {
		const off = EventsOn("session:completed", () => {
			setRunning(false);
		});
		return off;
	}, []);

	// Subscribe to Capture Session errors (issue #11). The session aborts on
	// the first collaborator failure; Go maps the cause to a human-readable
	// message which we surface as a dismissable red banner on the bar.
	useEffect(() => {
		const off = EventsOn("session:error", (data: unknown) => {
			const p = data as { message?: string };
			setRunning(false);
			setErrorMessage(
				p?.message ??
					"Something went wrong during the session. Please try again.",
			);
		});
		return off;
	}, []);

	useEffect(() => {
		if (!selectingRegion) return;
		const onKey = (e: KeyboardEvent) => {
			if (e.key === "Escape") cancelRegionSelection();
		};
		window.addEventListener("keydown", onKey);
		return () => window.removeEventListener("keydown", onKey);
	}, [selectingRegion, cancelRegionSelection]);

	async function chooseFolder() {
		const chosen = await ChooseOutputDirectory();
		if (chosen) {
			setOutputDir(chosen);
		}
	}

	async function beginRegionSelection() {
		const [size, pos] = await Promise.all([
			WindowGetSize(),
			WindowGetPosition(),
		]);
		originalWindowRef.current = { size, pos };
		// Relax the size lock so the user can resize the region-selection
		// window freely. The bar's Min/Max are re-applied in restoreWindow.
		WindowSetMinSize(200, 150);
		WindowSetMaxSize(0, 0);
		WindowSetSize(REGION_FRAME_WIDTH, REGION_FRAME_HEIGHT);
		setSelectingRegion(true);
	}

	async function confirmRegionSelection() {
		// The Go side reads the real NSWindow frame via cgo/Cocoa and
		// converts it to the primary-top-left, points coordinate space
		// that kbinani/screenshot.Capture expects. Doing this in JS via
		// WindowGetPosition + devicePixelRatio breaks on multi-display:
		// Wails returns *screen-local* coords, so a window on a secondary
		// display looks like the same offset on the primary display.
		//
		// The advance click point is derived from the marker's position
		// inside the transparent frame. Because the frame fills the whole
		// (frameless) window and CSS pixels equal points on macOS, the
		// screen-space click point is simply region.min + marker offset.
		try {
			const region = await GetSelectedRegion();
			setRegion({
				x: region.x,
				y: region.y,
				width: region.width,
				height: region.height,
			});
			setClickPoint({
				x: region.x + Math.round(markerPos.x),
				y: region.y + Math.round(markerPos.y),
			});
		} catch (e) {
			setStatus(`Failed to read window rect: ${String(e)}`);
		} finally {
			restoreWindow();
			setSelectingRegion(false);
		}
	}

	const parsedRepeatCount = Number.parseInt(repeatCount, 10);
	const repeatCountValid =
		Number.isInteger(parsedRepeatCount) && parsedRepeatCount >= 1;

	const parsedStepInterval = Number.parseFloat(stepInterval);
	const stepIntervalValid =
		Number.isFinite(parsedStepInterval) && parsedStepInterval > 0;

	const outputsValid = outputDir !== "" && outputFileName.trim() !== "";

	async function runTestSession() {
		if (!region || !clickPoint) return;
		setErrorMessage(null);
		setRunning(true);
		setStatus("Running…");
		try {
			// Go owns the Output Document path: it resolves name collisions by
			// appending "-2", "-3", ..., so the file written may not be the one
			// we asked for. Render what it returns; never re-assemble it here.
			const savedPath = await RunTestSession(
				new main.TestSessionParams({
					repeatCount: parsedRepeatCount,
					stepIntervalSeconds: parsedStepInterval,
					outputDir,
					outputFileName: outputFileName.trim(),
					captureRegion: region,
					advanceClickPoint: clickPoint,
				}),
			);
			setStatus(`Saved to ${savedPath}`);
		} catch {
			// The Go side emits a session:error event with a human-readable
			// message, which drives the red banner; nothing to do here beyond
			// swallowing the rejected promise.
		} finally {
			setRunning(false);
		}
	}

	function stopTestSession() {
		// Cooperative stop: the current Capture Step finishes, then the loop
		// ends and the Output Document is saved. The Go side emits
		// session:completed, which drives the bar to its finished state.
		StopSession();
	}

	return (
		<div id="App">
			{!selectingRegion && (
				<div className="floating-bar">
					{errorMessage ? (
						<div id="result" className="result result-error" role="alert">
							<span className="error-message">{errorMessage}</span>
							<button
								type="button"
								className="error-close"
								aria-label="Dismiss"
								onClick={() => setErrorMessage(null)}
							>
								×
							</button>
						</div>
					) : (
						<div id="result" className="result" title={status || undefined}>
							{running && (
								<span className="result-spinner" aria-hidden="true" />
							)}
							{/* text-overflow needs a block container; the flex row is not one. */}
							{status && <span className="result-text">{status}</span>}
						</div>
					)}
					<div
						id="input"
						className="input-box"
						role="toolbar"
						aria-label="pasha controls"
					>
						<div className="field">
							<label htmlFor="repeat-count">Repeat Count</label>
							<input
								id="repeat-count"
								type="number"
								className="input"
								min={1}
								value={repeatCount}
								onChange={(e) => setRepeatCount(e.target.value)}
							/>
						</div>
						<div className="field">
							<label htmlFor="step-interval">Step Interval (sec)</label>
							<input
								id="step-interval"
								type="number"
								className="input"
								min={0.1}
								step={0.1}
								value={stepInterval}
								onChange={(e) => setStepInterval(e.target.value)}
							/>
						</div>
						<div className="field">
							<label htmlFor="output-file-name">File Name</label>
							<input
								id="output-file-name"
								type="text"
								className="input"
								value={outputFileName}
								onChange={(e) => setOutputFileName(e.target.value)}
							/>
						</div>

						<div className="field-separator" aria-hidden="true" />

						<div className="field">
							<button type="button" className="btn" onClick={chooseFolder}>
								Choose Folder
							</button>
							{/* Read-only input, not a span: it ellipsises while unfocused yet
							    lets the caret walk the whole path once focused. */}
							<input
								type="text"
								readOnly
								aria-label="Output folder"
								className={`path-input${outputDir ? " path-input-rtl" : ""}`}
								title={outputDir || undefined}
								value={outputDir ? displayPath(outputDir) : ""}
								placeholder="(no folder chosen)"
							/>
						</div>

						<div className="field">
							<button
								type="button"
								className="btn"
								onClick={beginRegionSelection}
							>
								Set Range
							</button>
							{/* The Capture Region and the Advance Click Point are always
							    set together by confirmRegionSelection, so one indicator
							    covers both. The exact numbers are not actionable here —
							    the user verifies the framing in the selection window. */}
							<span
								role="status"
								className={`status-chip${region ? " status-chip-set" : ""}`}
							>
								<span className="status-chip-icon" aria-hidden="true">
									{region ? "✓" : "–"}
								</span>
								{region ? "Set" : "Not set"}
							</span>
						</div>

						<div className="field-spacer" />
						{/* Start and stop are mutually exclusive: the bar is too
						    narrow to show both, so a running session only shows Stop (issue #09). */}
						{running ? (
							<button
								type="button"
								className="btn btn-stop"
								onClick={stopTestSession}
							>
								Stop
							</button>
						) : (
							<button
								type="button"
								className="btn btn-primary"
								onClick={runTestSession}
								disabled={
									!repeatCountValid ||
									!stepIntervalValid ||
									!outputsValid ||
									!region ||
									!clickPoint
								}
							>
								Pasha
							</button>
						)}
					</div>
				</div>
			)}
			{selectingRegion && (
				<div
					role="dialog"
					aria-label="Capture Region selection"
					className="region-frame"
				>
					<button
						type="button"
						aria-label="Advance click point marker"
						className="click-point-marker"
						style={{ left: `${markerPos.x}px`, top: `${markerPos.y}px` }}
						onPointerDown={handleMarkerPointerDown}
						onPointerMove={handleMarkerPointerMove}
						onPointerUp={handleMarkerPointerUp}
					/>
					<div className="region-frame-toolbar">
						<button
							type="button"
							className="btn btn-primary"
							onClick={confirmRegionSelection}
						>
							Set
						</button>
						<button
							type="button"
							className="btn"
							onClick={cancelRegionSelection}
						>
							Cancel
						</button>
					</div>
				</div>
			)}
		</div>
	);
}

export default App;
