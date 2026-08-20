/// <reference types="astro/client" />

interface ImportMetaEnv {
  readonly PUBLIC_LOOPS_FORM_ID?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
