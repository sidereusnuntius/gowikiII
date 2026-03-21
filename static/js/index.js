document.addEventListener("fx:after", (event) => {
  if (!event || !event.detail) {
    return;
  }

  console.log("Response status:", event.detail.cfg.response.status);
  const res = event.detail.cfg.response;
  if (!res || !res.headers) {
    return;
  }

  const newURL = res.headers["FX-url"];
  if (newURL) {
  }

  console.log("headers:", res.headers);

  const selector = res.headers.get("fx-target");
  for (const x of res.headers) {
    console.log(x);
  }
  console.log("selector:", selector);

  // If a target is provided in the response via the fx-target header, then swap the element matching the selector in the header.
  if (selector) {
    event.detail.cfg.target = document.querySelector(selector);
    console.log("selector:", selector);
  }

  // Handling different error codes.
  // 303: trigger a fixi request to the path provided in the fx-redirect header. Using the default Location header was causing some problems I haven't managed to fix yet.
  // 400 Bad Request: validation errors.
  // >= 401: swap the error-display element. I think a better approach would be to have multiple headers with a common prefix, allowing the frontend to render errors on
  // multiple input fields, for instance.
  if (res.status == 303) {
    // McGyverism alert: since there is no way to programmatically trigger a fixi request, I have a "redirector" button. If we want to do a request from Javascript,
    // we set the redirector's attributes defined by the fixi library and click on the button.
    const button = document.querySelector("#redirector");
    button.setAttribute("fx-action", res.headers.get("fx-redirect"));
    button.setAttribute("fx-target", res.headers.get("fx-redirect-target"));
    console.log("should click on button");
    button.click();
  } else if (res.status >= 400) {
    event.detail.cfg.target = document.querySelector(".error-display");
    event.detail.cfg.swap = "innerHTML";
  }
  console.log("received FX response:", res.status);
  console.log(event);
});

console.log("hello from the other side! :)");
