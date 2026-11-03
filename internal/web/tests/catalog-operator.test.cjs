const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

const script = fs.readFileSync(path.join(__dirname, "../static/js/catalog-operator.js"), "utf8");

function field(value = "") {
    const events = new Map();
    return {
        value,
        addEventListener(name, callback) {
            events.set(name, callback);
        },
        input(value) {
            this.value = value;
            events.get("input")();
        },
    };
}

function creationForm(initialSlug = "") {
    const title = field();
    const slug = field(initialSlug);
    const form = {
        querySelector(selector) {
            return selector === "[name=title]" ? title : slug;
        },
    };
    vm.runInNewContext(script, {
        document: {
            addEventListener(name, callback) {
                assert.equal(name, "DOMContentLoaded");
                callback();
            },
            querySelector() {
                return form;
            },
        },
    });
    return { title, slug };
}

test("ASCII slug suggestions follow the server's generator contract", () => {
    const form = creationForm();
    for (const [title, expected] of [
        [" Pride & Prejudice! ", "pride-prejudice"],
        ["A---B 123", "a-b-123"],
        ["Книга", ""],
        ["Café", "caf"],
        ["İ K A", "a"],
        ["", ""],
    ]) {
        form.title.input(title);
        assert.equal(form.slug.value, expected);
    }
});

test("manual slug edits survive later title changes", () => {
    const form = creationForm();
    form.title.input("Original Title");
    form.slug.input("chosen-slug");
    form.title.input("Changed Title");
    assert.equal(form.slug.value, "chosen-slug");

    form.slug.input("");
    form.title.input("Another Title");
    assert.equal(form.slug.value, "");
});

test("a submitted slug is retained after server validation", () => {
    const form = creationForm("submitted-slug");
    form.title.input("Changed Title");
    assert.equal(form.slug.value, "submitted-slug");
});

test("pages without a creation form are left unchanged", () => {
    vm.runInNewContext(script, {
        document: {
            addEventListener(name, callback) {
                callback();
            },
            querySelector() {
                return null;
            },
        },
    });
});
