function onChangeInputHandler(event) {
  console.log(event)
  if (!event.target) {
    return;
  }
  const input = event.target;
  input.removeAttribute('aria-invalid');
  document.querySelector(`.error-msg[for=${input.id}]`).textContent = '';
}

function addEventListeners() {
  document.querySelectorAll('input').forEach((input) => {
    input.addEventListener('input', (event) => onChangeInputHandler(event))
  })
}

function handleCode(event) {
  const res = event.detail.cfg.response;
  // Handling different error codes.
  // 303: trigger a fixi request to the path provided in the fx-redirect header. Using the default Location header was causing some problems I haven't managed to fix yet.
  switch (res.status) {
    case 303:
      // McGyverism alert: since there is no way to programmatically trigger a fixi request, I have a "redirector" button. If we want to do a request from Javascript,
      // we set the redirector's attributes defined by the fixi library and click on the button.
      const button = document.querySelector("#redirector");
      button.setAttribute("fx-action", res.headers.get("fx-redirect"));
      button.setAttribute("fx-target", res.headers.get("fx-redirect-target"));
      button.click();
    case 400: // Bad request
      /*
      One or more input values provided by the user are incorrect. We expect the response to contain a set of headers, each starting with Err- followed by
      the name attribute of the incorrect input control. We use the value of each header as the error message for the corresponding input control
      */
      res.headers.forEach((value, name) => {
        if (name.length >= 4 && name.startsWith("err-")) {
          const fieldname = name.slice(4);
          const input = document.querySelector(`[name=${fieldname}]`);
          if (input) {
            input.setAttribute("aria-invalid", "true");
            input.setCustomValidity(value);
            document.querySelector(`.error-msg[for=${input.id}]`).textContent = value;
          }
        }
      });
      console.log(res.headers);
    default:
      const notice = document.querySelector(".error-message");
      if (notice) {
        notice.textContent = res.headers.get("fx-error");
        notice.focus();
      }
  }
  event.preventDefault();
}

document.addEventListener("fx:after", (event) => {
  addEventListeners();
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

  if (res.status >= 300) {
    handleCode(event);
  }
  console.log("received FX response:", res.status);
  console.log(event);
});

console.log("hello from the other side! :)");
addEventListeners()
