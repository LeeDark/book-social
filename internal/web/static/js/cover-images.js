(() => {
    const selector = "img[data-cover-image]";

    function hideFailedImage(image) {
        if (image.complete && image.naturalWidth === 0) {
            image.hidden = true;
        }
    }

    function checkImages(root) {
        if (root.matches?.(selector)) {
            hideFailedImage(root);
        }

        root.querySelectorAll(selector).forEach(hideFailedImage);
    }

    // Image errors do not bubble. Capture also covers images inserted by HTMX.
    document.addEventListener("error", (event) => {
        if (event.target?.matches?.(selector)) {
            event.target.hidden = true;
        }
    }, true);

    // Cached failures may have happened before this script or before insertion.
    document.addEventListener("DOMContentLoaded", () => checkImages(document));
    document.addEventListener("htmx:load", (event) => checkImages(event.target));
})();
