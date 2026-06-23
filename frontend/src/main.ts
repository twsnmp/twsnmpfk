import { mount } from "svelte";
import "./app.postcss";
import "@mdi/font/css/materialdesignicons.css"
import App from "./App.svelte";

if (typeof window !== "undefined") {
  const w = window as any;
  const logJS = (msg: string) => {
    if (w.go && w.go.main && w.go.main.App && w.go.main.App.LogFromJS) {
      w.go.main.App.LogFromJS(msg);
    }
  };

  w.logJS = logJS;

  window.onerror = function (message, source, lineno, colno, error) {
    const errInfo = `JS Error: ${message}\nSource: ${source}\nLine: ${lineno}:${colno}\nStack: ${error?.stack || ""}`;
    console.error("GLOBAL ERROR CAPTURED:", errInfo, error);
    logJS(errInfo);
    alert(errInfo);
    return false;
  };

  window.addEventListener("unhandledrejection", function (event) {
    const errInfo = `Unhandled Rejection: ${event.reason}\nStack: ${event.reason?.stack || ""}`;
    console.error("GLOBAL REJECTION CAPTURED:", errInfo, event.reason);
    logJS(errInfo);
    alert(errInfo);
  });

  // Intercept console.error to log to Go
  const originalConsoleError = console.error;
  console.error = function (...args) {
    originalConsoleError.apply(console, args);
    const msg = args.map(a => {
      try {
        return typeof a === "object" ? JSON.stringify(a) : String(a);
      } catch (e) {
        return String(a);
      }
    }).join(" ");
    logJS(`[console.error] ${msg}`);
  };
}

const app = mount(App, {
  target: document.getElementById("app") || document.body,
});

export default app;

