/* eslint-disable @typescript-eslint/no-require-imports */
const test = require("node:test");
const assert = require("node:assert/strict");

global.window = {};
require("./static/vip.js");

test("concierge writes streamed tokens through the reactive messages array", async () => {
  const state = window.vipConcierge();
  const messages = [];
  messages.push = function (...items) {
    // Alpine wraps inserted objects in proxies. Cloning here reproduces the
    // distinction between the original object and the reactive array entry.
    return Array.prototype.push.apply(this, items.map((item) => ({ ...item })));
  };
  state.messages = messages;
  state.input = "What has Karo shipped?";
  state.$refs = { transcript: { scrollTop: 0, scrollHeight: 100 } };
  state.$nextTick = (callback) => callback();

  const stream = [
    'event: token\ndata: "A grounded"\n\n',
    'event: token\ndata: " answer."\n\n',
    'event: done\ndata: {"state":"complete"}\n\n'
  ].join("");
  global.fetch = async () => ({
    ok: true,
    body: new ReadableStream({
      start(controller) {
        controller.enqueue(new TextEncoder().encode(stream));
        controller.close();
      }
    })
  });

  await state.submit();

  assert.equal(state.messages[1].content, "A grounded answer.");
  assert.equal(state.stateMessage, "Answer complete.");
  assert.equal(state.streaming, false);
});

test("concierge rejects a done event with no visible answer", async () => {
  const state = window.vipConcierge();
  state.input = "Question";
  state.$refs = { transcript: { scrollTop: 0, scrollHeight: 0 } };
  state.$nextTick = (callback) => callback();
  global.fetch = async () => ({
    ok: true,
    body: new ReadableStream({
      start(controller) {
        controller.enqueue(new TextEncoder().encode('event: done\ndata: {"state":"complete"}\n\n'));
        controller.close();
      }
    })
  });

  await state.submit();

  assert.equal(state.messages.length, 1);
  assert.match(state.stateMessage, /empty answer/i);
});
