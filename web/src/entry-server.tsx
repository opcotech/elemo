import { RouterProvider } from "@tanstack/react-router";
import { StrictMode } from "react";
import { renderToReadableStream } from "react-dom/server";

import { createRouter } from "./router";

function parseCookie(header = "") {
  return header.split(";").reduce<Record<string, string>>((acc, part) => {
    const [k, ...v] = part.trim().split("=");
    if (!k) return acc;
    acc[decodeURIComponent(k)] = decodeURIComponent(v.join("="));
    return acc;
  }, {});
}

export async function render(opts: {
  url: string;
  req: Request;
  head: string;
}) {
  const cookies = parseCookie(opts.req.headers.get("cookie") ?? "");

  const router = createRouter({
    context: {
      request: opts.req,
      accessToken: cookies.elemo_at,
      refreshToken: cookies.elemo_rt,
    },
  });

  await router.load();

  return renderToReadableStream(
    <StrictMode>
      <RouterProvider router={router} />
    </StrictMode>
  );
}
