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

function App() {
	const [status, setStatus] = useState(
		"Press the button to run a test session",
	);
	const [running, setRunning] = useState(false);
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
	// so we render "N / M ステップ完了" as text (issue #08).
	useEffect(() => {
		const off = EventsOn("session:progress", (data: unknown) => {
			const p = data as { current?: number; total?: number };
			if (typeof p?.current === "number" && typeof p?.total === "number") {
				setStatus(`${p.current} / ${p.total} ステップ完了`);
			}
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
		setRunning(true);
		setStatus("Running test session…");
		try {
			await RunTestSession(
				new main.TestSessionParams({
					repeatCount: parsedRepeatCount,
					stepIntervalSeconds: parsedStepInterval,
					outputDir,
					outputFileName: outputFileName.trim(),
					captureRegion: region,
					advanceClickPoint: clickPoint,
				}),
			);
			setStatus(`Done. Check ${outputDir}/${outputFileName.trim()}.pdf`);
		} catch (e) {
			setStatus(`Failed: ${String(e)}`);
		} finally {
			setRunning(false);
		}
	}

	return (
		<div id="App">
			{!selectingRegion && (
				<div className="floating-bar">
					<div id="result" className="result">
						{status}
					</div>
					<div
						id="input"
						className="input-box"
						role="toolbar"
						aria-label="pasha controls"
					>
						<label htmlFor="repeat-count">Repeat Count</label>
						<input
							id="repeat-count"
							type="number"
							className="input"
							min={1}
							value={repeatCount}
							onChange={(e) => setRepeatCount(e.target.value)}
						/>
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
						<label htmlFor="output-file-name">File Name</label>
						<input
							id="output-file-name"
							type="text"
							className="input"
							value={outputFileName}
							onChange={(e) => setOutputFileName(e.target.value)}
						/>
						<button type="button" className="btn" onClick={chooseFolder}>
							Choose Folder
						</button>
						<span className="output-dir">
							{outputDir || "(no folder chosen)"}
						</span>
						<button
							type="button"
							className="btn"
							onClick={beginRegionSelection}
						>
							範囲選択
						</button>
						<span className="region-indicator">
							{region
								? `範囲指定済み (${region.x},${region.y}) ${region.width}×${region.height}`
								: "(未指定)"}
						</span>
						<span className="click-point-indicator">
							{clickPoint
								? `クリック位置指定済み (${clickPoint.x},${clickPoint.y})`
								: ""}
						</span>
						<button
							type="button"
							className="btn"
							onClick={runTestSession}
							disabled={
								running ||
								!repeatCountValid ||
								!stepIntervalValid ||
								!outputsValid ||
								!region ||
								!clickPoint
							}
						>
							テスト撮影
						</button>
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
						aria-label="クリック位置マーカー"
						className="click-point-marker"
						style={{ left: `${markerPos.x}px`, top: `${markerPos.y}px` }}
						onPointerDown={handleMarkerPointerDown}
						onPointerMove={handleMarkerPointerMove}
						onPointerUp={handleMarkerPointerUp}
					/>
					<div className="region-frame-toolbar">
						<span className="region-frame-hint">
							この窓の範囲をキャプチャします。窓を移動・リサイズして位置を合わせてください。マーカーをドラッグしてクリック位置を決めてください。
						</span>
						<button
							type="button"
							className="btn btn-primary"
							onClick={confirmRegionSelection}
						>
							確定
						</button>
						<button
							type="button"
							className="btn"
							onClick={cancelRegionSelection}
						>
							キャンセル
						</button>
					</div>
				</div>
			)}
		</div>
	);
}

export default App;
