import logo from "./assets/images/logo-universal.png";
import "./App.css";
import { useState } from "react";
import { RunTestSession } from "../wailsjs/go/main/App";

function App() {
	const [status, setStatus] = useState(
		"Press the button to run a test session",
	);
	const [running, setRunning] = useState(false);

	async function runTestSession() {
		setRunning(true);
		setStatus("Running test session…");
		try {
			await RunTestSession();
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
				<button
					type="button"
					className="btn"
					onClick={runTestSession}
					disabled={running}
				>
					テスト撮影
				</button>
			</div>
		</div>
	);
}

export default App;
