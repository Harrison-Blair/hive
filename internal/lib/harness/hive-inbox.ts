// hive-inbox: the native message path of hive agent send for pi. hive writes
// one JSON line {"text": "⬢ SENDER: TEXT"} to $XDG_RUNTIME_DIR/hive-pi/<pid>.sock.
import * as fs from "node:fs";
import * as net from "node:net";
import * as os from "node:os";
import * as path from "node:path";
import type { ExtensionAPI, ExtensionContext } from "@earendil-works/pi-coding-agent";

export default function (pi: ExtensionAPI) {
	const dir = path.join(process.env.XDG_RUNTIME_DIR || os.tmpdir(), "hive-pi");
	const socket = path.join(dir, `${process.pid}.sock`);
	let ctx: ExtensionContext | undefined;
	let server: net.Server | undefined;
	const deliver = (line: string) => {
		const text = JSON.parse(line).text;
		if (ctx?.isIdle()) pi.sendUserMessage(text);
		else pi.sendUserMessage(text, { deliverAs: "followUp" });
	};
	pi.on("session_start", (_event, context) => {
		ctx = context;
		if (server) return;
		fs.mkdirSync(dir, { recursive: true, mode: 0o700 });
		fs.rmSync(socket, { force: true });
		server = net.createServer((conn) => {
			let data = "";
			conn.on("data", (chunk) => (data += chunk));
			conn.on("end", () => data.split("\n").filter(Boolean).forEach(deliver));
		});
		server.listen(socket);
	});
	pi.on("session_shutdown", () => {
		server?.close();
		server = undefined;
		fs.rmSync(socket, { force: true });
	});
}
