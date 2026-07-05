import logo from "./assets/images/logo-universal.png";
import "./App.css";
import { useEffect, useState } from "react";
import {
	ChooseOutputDirectory,
	DefaultOutputFileName,
	RunTestSession,
} from "../wailsjs/go/main/App";

function App() {
	const [status, setStatus] = useState(
		"Press the button to run a test session",
	);
	const [running, setRunning] = useState(false);
	const [repeatCount, setRepeatCount] = useState("10");
	const [stepInterval, setStepInterval] = useState("1.0");
	const [outputFileName, setOutputFileName] = useState("");
	const [outputDir, setOutputDir] = useState("");

	useEffect(() => {
		DefaultOutputFileName().then(setOutputFileName);
	}, []);

	async function chooseFolder() {
		const chosen = await ChooseOutputDirectory();
		if (chosen) {
			setOutputDir(chosen);
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
		setRunning(true);
		setStatus("Running test session…");
		try {
			await RunTestSession({
				repeatCount: parsedRepeatCount,
				stepIntervalSeconds: parsedStepInterval,
				outputDir,
				outputFileName: outputFileName.trim(),
			});
			setStatus(`Done. Check ${outputDir}/${outputFileName.trim()}.pdf`);
		} catch (e) {
			setStatus(`Failed: ${String(e)}`);
		} finally {
			setRunning(false);
		}
	}

	return (
		<div id="App">
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
				<span className="output-dir">{outputDir || "(no folder chosen)"}</span>
				<button
					type="button"
					className="btn"
					onClick={runTestSession}
					disabled={
						running || !repeatCountValid || !stepIntervalValid || !outputsValid
					}
				>
					テスト撮影
				</button>
			</div>
		</div>
	);
}

export default App;
