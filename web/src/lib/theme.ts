import { createServerFn } from "@tanstack/react-start";
import { getCookie, setCookie } from "@tanstack/react-start/server";
import { z } from "zod";

export const themeSchema = z.enum(["light", "dark", "system"]);
export type Theme = z.infer<typeof themeSchema>;

const THEME_COOKIE = "elemo_theme";

function themeCookieSecure(): boolean {
  const explicit = process.env.ELEMO_COOKIE_SECURE;
  if (explicit === "true") return true;
  if (explicit === "false") return false;
  return (process.env.APP_URL ?? "").startsWith("https://");
}

export const getThemeFn = createServerFn({ method: "GET" }).handler(() => {
  const parsed = themeSchema.safeParse(getCookie(THEME_COOKIE));
  return parsed.success ? parsed.data : "system";
});

export const setThemeFn = createServerFn({ method: "POST" })
  .validator(themeSchema)
  .handler(({ data }) => {
    setCookie(THEME_COOKIE, data, {
      httpOnly: true,
      secure: themeCookieSecure(),
      sameSite: "lax",
      path: "/",
      maxAge: 365 * 24 * 60 * 60,
    });
    return data;
  });
