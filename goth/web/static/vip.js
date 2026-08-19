(function () {
  "use strict";
  window.vipConcierge = function () {
    return {
      messages: [], input: "", streaming: false, stateMessage: "", controller: null, nextID: 1,
      ask: function (text) { this.input = text; this.submit(); },
      clear: function () { this.stop(); this.messages = []; this.stateMessage = "Conversation cleared."; },
      stop: function () { if (this.controller) this.controller.abort(); this.controller = null; this.streaming = false; },
      submit: async function () {
        var text = this.input.trim(); if (!text || this.streaming) return;
        this.input = ""; this.messages.push({ id: this.nextID++, role: "user", content: text });
        var answerIndex = this.messages.length;
        this.messages.push({ id: this.nextID++, role: "assistant", content: "" });
        this.streaming = true; this.stateMessage = "AI guide is thinking..."; this.controller = new AbortController();
        try {
          var response = await fetch("/api/vip/chat", { method: "POST", credentials: "same-origin", headers: { "Content-Type": "application/json", Accept: "text/event-stream" }, body: JSON.stringify({ message: text }), signal: this.controller.signal });
          if (!response.ok || !response.body) throw new Error(response.status === 429 ? "The concierge is taking a short break. Please try again later." : "The concierge is temporarily unavailable.");
          var reader = response.body.getReader(), decoder = new TextDecoder(), buffer = "";
          while (true) {
            var part = await reader.read(); if (part.done) break; buffer += decoder.decode(part.value, { stream: true });
            var frames = buffer.split("\n\n"); buffer = frames.pop();
            for (var i = 0; i < frames.length; i++) {
              var lines = frames[i].split("\n"), event = "message", data = "";
              for (var j = 0; j < lines.length; j++) { if (lines[j].indexOf("event: ") === 0) event = lines[j].slice(7); if (lines[j].indexOf("data: ") === 0) data += lines[j].slice(6); }
              if (!data) continue; var parsed = JSON.parse(data);
              if (event === "token" && typeof parsed === "string") this.messages[answerIndex].content += parsed;
              if (event === "error") throw new Error(parsed.error || "The concierge could not complete that answer.");
              if (event === "done") {
                if (!this.messages[answerIndex].content.trim()) throw new Error("The concierge returned an empty answer. Please try again.");
                this.stateMessage = "Answer complete.";
              }
            }
            this.$nextTick(function () { this.$refs.transcript.scrollTop = this.$refs.transcript.scrollHeight; }.bind(this));
          }
        } catch (error) { if (error.name === "AbortError") this.stateMessage = "Generation stopped."; else { this.stateMessage = error.message; this.messages.splice(answerIndex, 1); } }
        finally { this.controller = null; this.streaming = false; }
      }
    };
  };
}());
