import logo from "./assets/images/logo-universal.png";
import "./App.css";
import { useState } from "react";
import { RunTestSession } from "../wailsjs/go/main/App";

function App() {
	const [status, setStatus] = useState(
		"Press the button to run a test session",
	);
	const [running, setRunning] = useState(false);
	const [repeatCount, setRepeatCount] = useState("10");
	const [stepInterval, setStepInterval] = useState("1.0");

	const parsedRepeatCount = Number.parseInt(repeatCount, 10);
	const repeatCountValid =
		Number.isInteger(parsedRepeatCount) && parsedRepeatCount >= 1;

	const parsedStepInterval = Number.parseFloat(stepInterval);
	const stepIntervalValid =
		Number.isFinite(parsedStepInterval) && parsedStepInterval > 0;

	async function runTestSession() {
		setRunning(true);
		setStatus("Running test session…");
		try {
			await RunTestSession(parsedRepeatCount, parsedStepInterval);
			setStatus("Done. Check ~/Desktop/pasha-tracer.pdf");
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
				<button
					type="button"
					className="btn"
					onClick={runTestSession}
					disabled={running || !repeatCountValid || !stepIntervalValid}
				>
					テスト撮影
				</button>
			</div>
		</div>
	);
}

export default App;
