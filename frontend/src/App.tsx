import logo from "./assets/images/logo-universal.png";
import "./App.css";
import { useCallback, useEffect, useRef, useState } from "react";
import {
	ChooseOutputDirectory,
	DefaultOutputFileName,
	RunTestSession,
} from "../wailsjs/go/main/App";
import { main } from "../wailsjs/go/models";
import {
	ScreenGetAll,
	WindowGetPosition,
	WindowGetSize,
	WindowSetAlwaysOnTop,
	WindowSetPosition,
	WindowSetSize,
} from "../wailsjs/runtime/runtime";

type CaptureRegion = {
	x: number;
	y: number;
	width: number;
	height: number;
};

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
	const [rubberBand, setRubberBand] = useState<{
		x: number;
		y: number;
		width: number;
		height: number;
	} | null>(null);
	const dragStartRef = useRef<{
		screen: { x: number; y: number };
		client: { x: number; y: number };
	} | null>(null);
	const originalWindowRef = useRef<{
		size: { w: number; h: number };
		pos: { x: number; y: number };
	} | null>(null);

	const restoreWindow = useCallback(() => {
		const orig = originalWindowRef.current;
		WindowSetAlwaysOnTop(false);
		if (orig) {
			WindowSetPosition(orig.pos.x, orig.pos.y);
			WindowSetSize(orig.size.w, orig.size.h);
		}
	}, []);

	useEffect(() => {
		DefaultOutputFileName().then(setOutputFileName);
	}, []);

	useEffect(() => {
		if (!selectingRegion) return;
		const onKey = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				dragStartRef.current = null;
				setRubberBand(null);
				restoreWindow();
				setSelectingRegion(false);
			}
		};
		window.addEventListener("keydown", onKey);
		return () => window.removeEventListener("keydown", onKey);
	}, [selectingRegion, restoreWindow]);

	async function chooseFolder() {
		const chosen = await ChooseOutputDirectory();
		if (chosen) {
			setOutputDir(chosen);
		}
	}

	async function beginRegionSelection() {
		const [size, pos, screens] = await Promise.all([
			WindowGetSize(),
			WindowGetPosition(),
			ScreenGetAll(),
		]);
		originalWindowRef.current = { size, pos };
		const primary = screens.find((s) => s.isPrimary) ?? screens[0];
		WindowSetAlwaysOnTop(true);
		WindowSetPosition(0, 0);
		WindowSetSize(primary.width, primary.height);
		setSelectingRegion(true);
	}

	function handleOverlayMouseDown(event: React.MouseEvent<HTMLDivElement>) {
		dragStartRef.current = {
			screen: { x: event.screenX, y: event.screenY },
			client: { x: event.clientX, y: event.clientY },
		};
		setRubberBand({ x: event.clientX, y: event.clientY, width: 0, height: 0 });
	}

	function handleOverlayMouseMove(event: React.MouseEvent<HTMLDivElement>) {
		const start = dragStartRef.current;
		if (!start) return;
		setRubberBand({
			x: Math.min(start.client.x, event.clientX),
			y: Math.min(start.client.y, event.clientY),
			width: Math.abs(event.clientX - start.client.x),
			height: Math.abs(event.clientY - start.client.y),
		});
	}

	function handleOverlayMouseUp(event: React.MouseEvent<HTMLDivElement>) {
		const start = dragStartRef.current;
		if (!start) return;
		const x = Math.min(start.screen.x, event.screenX);
		const y = Math.min(start.screen.y, event.screenY);
		const width = Math.abs(event.screenX - start.screen.x);
		const height = Math.abs(event.screenY - start.screen.y);
		dragStartRef.current = null;
		setRubberBand(null);
		if (width > 0 && height > 0) {
			setRegion({ x, y, width, height });
		}
		restoreWindow();
		setSelectingRegion(false);
	}

	const parsedRepeatCount = Number.parseInt(repeatCount, 10);
	const repeatCountValid =
		Number.isInteger(parsedRepeatCount) && parsedRepeatCount >= 1;

	const parsedStepInterval = Number.parseFloat(stepInterval);
	const stepIntervalValid =
		Number.isFinite(parsedStepInterval) && parsedStepInterval > 0;

	const outputsValid = outputDir !== "" && outputFileName.trim() !== "";

	async function runTestSession() {
		if (!region) return;
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
				<div className="main-panel">
					<img src={logo} id="logo" alt="logo" />
					<div id="result" className="result">
						{status}
					</div>
					<div id="input" className="input-box">
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
						<button
							type="button"
							className="btn"
							onClick={runTestSession}
							disabled={
								running ||
								!repeatCountValid ||
								!stepIntervalValid ||
								!outputsValid ||
								!region
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
					className="region-overlay"
					onMouseDown={handleOverlayMouseDown}
					onMouseMove={handleOverlayMouseMove}
					onMouseUp={handleOverlayMouseUp}
				>
					<div className="region-hint">
						ドラッグで範囲を選択 / Esc でキャンセル
					</div>
					{rubberBand && (
						<div
							className="region-rubber-band"
							style={{
								left: rubberBand.x,
								top: rubberBand.y,
								width: rubberBand.width,
								height: rubberBand.height,
							}}
						>
							<span className="region-rubber-band-label">
								{rubberBand.width} × {rubberBand.height}
							</span>
						</div>
					)}
				</div>
			)}
		</div>
	);
}

export default App;
