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
	// pi 1.4.0 rejects a message without deliverAs while a run starts or runs,
	// and every message during a manual compaction, and drops a rejected
	// message. So one message goes to an idle pi; the next ones wait until its
	// run starts and then go as follow-ups. While pi is neither idle nor in a
	// run, the queue waits and checks again every 250 ms.
	const queue: string[] = [];
	let running = false; // from agent_start to agent_settled
	let sentAt = 0; // when a message went to an idle pi whose run has not started yet
	let timer: ReturnType<typeof setTimeout> | undefined;
	const pump = () => {
		timer = undefined;
		try {
			while (queue.length > 0) {
				if (running) pi.sendUserMessage(queue.shift()!, { deliverAs: "followUp" });
				else if (Date.now() - sentAt < 10_000 || !ctx?.isIdle()) break;
				else {
					sentAt = Date.now();
					pi.sendUserMessage(queue.shift()!);
				}
			}
		} catch {
			// A socket callback never throws: an error there would end pi.
		}
		if (queue.length > 0 && !timer) timer = setTimeout(pump, 250);
	};
	const receive = (line: string) => {
		try {
			const text = JSON.parse(line).text;
			if (typeof text === "string") queue.push(text);
		} catch {
			// Skip a line that is not a hive message.
		}
	};
	pi.on("agent_start", () => {
		running = true;
		sentAt = 0;
		pump();
	});
	pi.on("agent_settled", () => {
		running = false;
		pump();
	});
	pi.on("session_start", (_event, context) => {
		ctx = context;
		if (server) return;
		fs.mkdirSync(dir, { recursive: true, mode: 0o700 });
		fs.rmSync(socket, { force: true });
		server = net.createServer((conn) => {
			let data = "";
			conn.setEncoding("utf8"); // keeps a character split across chunks
			conn.on("data", (chunk) => (data += chunk));
			conn.on("error", () => {});
			// Only a newline ends a record: drop a record cut before its newline.
			conn.on("end", () => {
				data.split("\n").slice(0, -1).filter(Boolean).forEach(receive);
				pump();
			});
		});
		server.listen(socket);
	});
	pi.on("session_shutdown", () => {
		clearTimeout(timer);
		server?.close();
		server = undefined;
		fs.rmSync(socket, { force: true });
	});
}
