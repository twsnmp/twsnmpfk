import "./app.postcss";
import "@mdi/font/css/materialdesignicons.css"
import App from "./App.svelte";

console.log("main");

const app = new App({
  target: document.getElementById("app"),
});

export default app;

