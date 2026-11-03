const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

const script = fs.readFileSync(path.join(__dirname, "../static/js/cover-images.js"), "utf8");

function image({ complete = true, naturalWidth = 0, isCover = true } = {}) {
    return {
        complete,
        naturalWidth,
        hidden: false,
        matches(selector) {
            assert.equal(selector, "img[data-cover-image]");
            return isCover;
        },
        querySelectorAll() {
            return [];
        },
    };
}

function boot(images = []) {
    const listeners = new Map();
    const document = {
        querySelectorAll() {
            return images;
        },
        addEventListener(name, callback, capture = false) {
            listeners.set(name, { callback, capture });
        },
    };
    vm.runInNewContext(script, { document });

    return {
        listeners,
        fire(name, target = document) {
            listeners.get(name).callback({ target });
        },
    };
}

test("cached failures reveal the placeholder on initial load", () => {
    const failed = image();
    const page = boot([failed]);
    page.fire("DOMContentLoaded");
    assert.equal(failed.hidden, true);
});

test("loaded and pending lazy images remain visible", () => {
    const loaded = image({ naturalWidth: 240 });
    const pending = image({ complete: false });
    const page = boot([loaded, pending]);
    page.fire("DOMContentLoaded");
    assert.equal(loaded.hidden, false);
    assert.equal(pending.hidden, false);
});

test("captured errors cover initial and dynamically inserted images", () => {
    const page = boot();
    assert.equal(page.listeners.get("error").capture, true);

    const failed = image({ complete: false });
    page.fire("error", failed);
    assert.equal(failed.hidden, true);
});

test("HTMX fragments reveal cached failures but retain successful images", () => {
    const failed = image();
    const loaded = image({ naturalWidth: 240 });
    const page = boot();
    page.fire("htmx:load", { querySelectorAll: () => [failed, loaded] });
    assert.equal(failed.hidden, true);
    assert.equal(loaded.hidden, false);
});

test("an image can itself be the HTMX load root", () => {
    const failed = image();
    boot().fire("htmx:load", failed);
    assert.equal(failed.hidden, true);
});

test("unrelated resource errors are left unchanged", () => {
    const unrelated = image({ isCover: false });
    const page = boot();
    page.fire("error", unrelated);
    page.fire("error", {});
    page.fire("error", null);
    assert.equal(unrelated.hidden, false);
});
