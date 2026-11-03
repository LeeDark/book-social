document.addEventListener("DOMContentLoaded", () => {
    const form = document.querySelector("[data-operator-new-form]");
    if (!form) {
        return;
    }

    const title = form.querySelector("[name=title]");
    const slug = form.querySelector("[name=slug]");
    // Retain a submitted proposal and stop suggesting after any manual edit.
    let slugWasEdited = slug.value !== "";

    slug.addEventListener("input", () => {
        slugWasEdited = true;
    });

    title.addEventListener("input", () => {
        if (slugWasEdited) {
            return;
        }
        // Only ASCII case conversion: Unicode characters remain separators.
        slug.value = title.value
            .replace(/[A-Z]/g, (letter) => letter.toLowerCase())
            .replace(/[^a-z0-9]+/g, "-")
            .replace(/^-+|-+$/g, "");
    });
});
