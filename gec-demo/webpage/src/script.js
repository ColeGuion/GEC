//const GEC_ENDPOINT = "/api/gec";
const GEC_ENDPOINT = "http://localhost:8089/api/gec"; 
const EXAMPLES = {
    ex1: `We should buy a car.\n\nin conclusion, brutus strongly though that caeser was a bad leader.`,
    ex2: `when i got to the store i seen my friend. She say she don't got no money, but she wanted to buy apples oranges and bananas`,
    ex3: `Just over two months later, an attempt on his life would be made and failed. He quickly recovered and returned to duty, and this caused his popularity to sky rocket.\n\nIn conclusion, Brutus strongly though that caeser was a bad leader, which is why he was apart of his death.`,
    sva: `The list of items are on the table. Each of the dogs bark loudly at night.`,
    apostrophe: `The dog chased it's tail. Its almost two oclock.\nThe article contradicted it's own argument.\nIts best to do some research before deciding on a topic.\n\nMany people make a tradition of going for a long walk on New Years Day to clear their heads for the months ahead.\nThe dogs bark was far worse than its bite.\nLittle girls clothing is on the first floor, and the mens department is on the second.`,
    punc: `After dinner we went for a walk it was cold outside however we stayed out anyway.`,
    commas: `The list of items, is on the table. We followed the list and bought apples peaches and bananas today.\n\nNear a small stream at the bottom of the canyon park rangers discovered a gold mine.\n\nMary promised that she would be a good girl that she would not bite her brother and that she would not climb onto the television.\n\nThe instructor looked through his briefcase through his desk and around the office for the lost grade book.\n\nSteven Smith whose show you like will host a party next week.`,
    caps: `yesterday i visited chicago and met dr. smith at o'hare airport.\n\nAn gift from france, the statue of liberty has welcomed immigrants and visitors to New york Harbor since 1886.`,
    pronouns: `The movie turned out to be a blockbuster hit, who came as a surprise to critics.`,
    homophones: `I except your invitation to the wedding.\n\nThey went on a hike to watch for dear in the forest.\n\nThe puppy gave my finger a playful byte.`,
    plural: `Studies are showing that man process information differently from women.\n\nI wishes I could grant all your wish.\n\nThe bus was running late, which meant all the other bus were as well.`,
    hyphens: `She jumped from a two story building.\n\nWe offer around the clock coverage.\n\nIf we split the bill evenly, we each owe thirty four dollars.`,
    misspellings: `I met Mathilde yesterday in Salt Lake City. They have the best bowling rink of all time imo.`,
    profane: `That is your anus.`,
};


document.addEventListener("DOMContentLoaded", function () {
    const editorEl = document.getElementById("editorEl");
    const wordCountEl = document.getElementById("wordCount");
    const spellCountEl = document.getElementById("spellCount");
    const grammarCountEl = document.getElementById("grammarCount");
    const btnCheck = document.getElementById("btnCheck");
    const btnReset = document.getElementById("btnReset");
    const btnCopy = document.getElementById("btnCopy");
    const exampleSelectEl = document.getElementById("exampleSelect");
    // Keep last raw plain text (so offsets always refer to the correct text)
    let lastPlainText = "";
    //DummyFill();


    // ---------- Helpers ----------
    function updateStats({ plainText, textMarkup } = {}) {
        const text = plainText ?? getPlainTextFromEditor();
        const markups = Array.isArray(textMarkup) ? textMarkup : [];

        const wordCount = countWords(text);
        const spelling = markups.filter(m => isSpellingCategory(m.category)).length;
        const grammar = markups.length - spelling;

        wordCountEl.textContent = String(wordCount);
        spellCountEl.textContent = String(spelling);
        grammarCountEl.textContent = String(grammar);
    }

    function setEditorPlainText(text) {
        // Reset contenteditable to plain text with paragraphs
        editorEl.innerHTML = "";
        const lines = String(text ?? "").split(/\n/);
        for (const line of lines) {
            const p = document.createElement("p");
            p.textContent = line || "\u00A0";
            editorEl.appendChild(p);
        }
        lastPlainText = getPlainTextFromEditor();
        updateStats({ plainText: lastPlainText, textMarkup: [] });
    }

    // ---------- Main action ----------
    async function runCheck() {
        const text = getPlainTextFromEditor();
        lastPlainText = text;

        if (!text.trim()) {
            updateStats({ plainText: text, textMarkup: [] });
            return;
        }

        btnCheck.disabled = true;
        btnCheck.textContent = "Checking…";

        try {
            const resp = await fetch(GEC_ENDPOINT, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ text })
            });
            if (!resp.ok) {
                const errText = await resp.text().catch(() => "");
                throw new Error(`HTTP ${resp.status} ${resp.statusText}${errText ? " — " + errText : ""}`);
            }

            // response format:
            // {
            //   corrected_text: string,
            //   text_markups: [{ index, length, message, category }]
            //   gibberish_scores: [{ index, length, score: { clean, mild, noise, wordSalad } }]
            //   character_count: int,
            //   error_character_count: int,
            //   contains_profanity: bool,
            //   service_time: float,
            // }
            const data = await resp.json();
            logDataDetails(data);

            const corrected = data?.corrected_text ?? "";
            const markups = Array.isArray(data?.text_markups) ? data.text_markups : [];

            // Apply highlights DIRECTLY into the editable area
            applyHighlightsToEditor(markups, text);

            // Render highlights based on OFFSETS into the ORIGINAL text that we sent.
            //let markedHtml = buildMarkedHtmlFromOffsets(text, markups);
            //console.log(`Marked HTML: ${markedHtml}`);
            //renderEl.innerHTML = markedHtml;

            // Update counts from the returned markups + current text
            updateStats({ plainText: text, textMarkup: markups });

            //showRender();
        } catch (err) {
            console.error(err);
            // Stay in editor view but still update word count
            //updateStats({ plainText: text, textMarkup: [] });
            alert(`Check failed: ${err?.message ?? String(err)}`);
            clearHighlights();           // important: clean up on error
            //showEditor();
        } finally {
            btnCheck.disabled = false;
            btnCheck.textContent = "Check";
        }
    }

    // ---------- Events ----------
    editorEl.addEventListener("input", () => {
        // Whenever user types → remove all highlights (simplest UX)
        clearHighlights();

        // Update stats to zero
        updateStats({ plainText: getPlainTextFromEditor(), textMarkup: [] });
    });

    btnCheck.addEventListener("click", runCheck);
    btnReset.addEventListener("click", () => {
        setEditorPlainText("");
        editorEl.focus();
    });
    btnCopy.addEventListener("click", async () => {
        try {
            await navigator.clipboard.writeText(getPlainTextFromEditor());
        } catch {
            // ignore
        }
    });

    exampleSelectEl.addEventListener("change", () => {
        const key = exampleSelectEl.value;
        if (!key) return;
        setEditorPlainText(EXAMPLES[key] || "");
        editorEl.focus();
    });

    
    function applyHighlightsToEditor(markups, originalText) {
        if (!markups.length) {
            clearHighlights();
            return;
        }

        // Sort and normalize
        const marks = normalizeMarkups(markups, originalText.length);

        // Rebuild the content with <span> wrappers
        let html = "";
        let cursor = 0;

        for (const m of marks) {
            if (cursor < m.index) {
                html += escapeHtml(originalText.slice(cursor, m.index));
            }

            const seg = originalText.slice(m.index, m.end);
            const cls = isSpellingCategory(m.category) ? "spell" : "hl";
            const tooltip = safeTooltip(m.message || m.category || "Suggestion");

            html += `<span class="${cls}" data-tooltip="${escapeHtml(tooltip)}">${escapeHtml(seg || " ")}</span>`;

            cursor = m.end;
        }

        if (cursor < originalText.length) {
            html += escapeHtml(originalText.slice(cursor));
        }

        // Preserve line breaks
        html = html.replaceAll("\n", "<br>");

        // Put it back into the editor
        editorEl.innerHTML = html;
    }

    function clearHighlights() {
        // Flatten back to plain text (removes all <span> wrappers)
        const plain = getPlainTextFromEditor();
        setEditorPlainText(plain);   // rebuilds <p> tags
    }


    // ---------- Helpers ----------
    function buildMarkedHtmlFromOffsets(originalText, textMarkup) {
        const t = originalText ?? "";
        const marks = normalizeMarkups(textMarkup, t.length);

        let out = "";
        let cursor = 0;

        for (const m of marks) {
            if (cursor < m.index) out += escapeHtml(t.slice(cursor, m.index));

            const seg = t.slice(m.index, m.end);
            const cls = isSpellingCategory(m.category) ? "spell" : "hl";
            const tooltip = safeTooltip(m.message || m.category || "Suggestion");
            out += `<span class="${cls}" data-tooltip="${escapeHtml(tooltip)}">${escapeHtml(seg || " ")}</span>`;

            cursor = m.end;
        }

        if (cursor < t.length) out += escapeHtml(t.slice(cursor));

        // Preserve newlines visually like the editor text
        return out.replaceAll("\n", "<br/>");
    }

    function normalizeMarkups(textMarkup, textLen) {
        const arr = Array.isArray(textMarkup) ? textMarkup : [];
        const norm = arr.map(m => ({
            index: Number(m.index),
            length: Number(m.length),
            message: m.message ?? "",
            category: m.category ?? ""
        }))
            .filter(m => Number.isFinite(m.index) && Number.isFinite(m.length))
            .filter(m => m.index >= 0 && m.length >= 0)
            .map(m => ({ ...m, end: m.index + m.length }))
            .filter(m => m.end <= textLen)
            .sort((a, b) => a.index - b.index || a.end - b.end);

        // Drop overlaps
        const cleaned = [];
        let lastEnd = -1;
        for (const m of norm) {
            if (m.index < lastEnd) continue;
            cleaned.push(m);
            lastEnd = m.end;
        }
        return cleaned;
    }

    function getPlainTextFromEditor() {
        // contenteditable -> innerText gives a plain view with newlines
        return document.getElementById("editorEl").innerText || "";
    }

    function isSpellingCategory(cat) {
        const c = String(cat || "").toLowerCase();
        return c.includes("spelling"); // "SPELLING_MISTAKE"
    }

    function countWords(text) {
        const t = (text || "").trim();
        if (!t) return 0;
        return t.split(/\s+/).filter(Boolean).length;
    }

    function escapeHtml(str) {
        return String(str)
            .replaceAll("&", "&amp;")
            .replaceAll("<", "&lt;")
            .replaceAll(">", "&gt;")
            .replaceAll('"', "&quot;")
            .replaceAll("'", "&#039;");
    }

    function safeTooltip(msg) {
        const s = String(msg ?? "").trim();
        return s.length > 400 ? (s.slice(0, 397) + "...") : s;
    }


    function logDataDetails(data) {
        console.log(`Response Object:`, data);

        const corrected = data?.corrected_text ?? "";
        const markups = Array.isArray(data?.text_markups) ? data.text_markups : [];
        const charCount = data?.character_count ?? -1;
        const errCharCount = data?.error_character_count ?? -1;
        const hasProfanity = data?.contains_profanity ?? false;
        const serviceTime = data?.service_time ?? 0.0;
        console.log(`Corrected: ${JSON.stringify(corrected)}
Markups: ${markups.length}
Service Time: ${serviceTime}
Character Count: ${charCount}
Error Characters: ${errCharCount}
Has Profanity?: ${hasProfanity}`);
        console.log("Markups:", markups);
    }

    // ---------- Init ----------
    setEditorPlainText("we shood buy an car.");
});

