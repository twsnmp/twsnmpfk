import { addMessages, init } from "svelte-i18n";
import en from "./en.json";
import ja from "./ja.json";
import {
  GetLang,
} from "../../wailsjs/go/main/App";

addMessages("en", en);
addMessages("ja", ja);
export const lang = await GetLang();
init({
    fallbackLocale: "en",
    initialLocale: lang || "en",
});
