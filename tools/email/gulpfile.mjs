import { rmSync, mkdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import minimist from 'minimist';
import { src, dest, series } from 'gulp';
import replace from 'gulp-replace';
import fileinclude from 'gulp-file-include';
import purgecss from 'gulp-purgecss';
import htmlmin from 'gulp-htmlmin';
import imagemin, { mozjpeg, optipng } from 'gulp-imagemin';
import inlineSource from 'gulp-inline-source';
import exec from 'gulp-exec';
import s3Uploader from '@opcotech/gulp-s3-upload';

const options = minimist(process.argv.slice(2));

const s3Client = s3Uploader(
  {
    accessKeyId: options?.['access-key-id'],
    secretAccessKey: options?.['secret-access-key'],
    region: options?.['region'],
  },
  {
    endpoint: options?.['endpoint'],
    forcePathStyle: true,
  }
);

const EMAIL_DIR = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(EMAIL_DIR, '../..');
const PREMAILER_DIR = path.join(EMAIL_DIR, 'pre-mailer');
const BUILD_DIR = path.join(EMAIL_DIR, 'build');
const SRC_DIR = path.join(EMAIL_DIR, 'src');
const DIST_DIR = options.out
  ? path.isAbsolute(options.out)
    ? options.out
    : path.resolve(REPO_ROOT, options.out)
  : path.join(REPO_ROOT, 'templates/email');

const IMAGE_FILES = `${SRC_DIR}/**/*.(png|jpg|jpeg)`;
const CSS_FILES = `${SRC_DIR}/**/*.css`;
const HTML_FILES = [`${SRC_DIR}/**/*.html`, `!${SRC_DIR}/includes/*.html`];

const BUCKET_STATIC_PATH = 'email-assets/';

function clean(cb) {
  rmSync(BUILD_DIR, { force: true, recursive: true });
  mkdirSync(BUILD_DIR);

  cb();
}

function minifyImages() {
  let stream = src(IMAGE_FILES, { encoding: false }).pipe(imagemin([mozjpeg(), optipng()], { silent: true }));

  if (options?.['s3-bucket']) {
    stream = stream.pipe(
      s3Client({
        Bucket: options['s3-bucket'],
        keyTransform: (key) => {
          return key.replace('assets/', BUCKET_STATIC_PATH);
        },
      })
    );
  }

  return stream.pipe(dest(BUILD_DIR));
}

function minifyCSS() {
  return src(CSS_FILES)
    .pipe(purgecss({ content: [HTML_FILES] }))
    .pipe(dest(BUILD_DIR));
}

function minifyHTML() {
  let stream = src(HTML_FILES)
    .pipe(
      fileinclude({
        prefix: '@@',
        basepath: '@file',
      })
    )
    .pipe(inlineSource());

  if (options?.['static-root']) {
    stream = stream.pipe(replace('./assets/', `${options['static-root']}/${BUCKET_STATIC_PATH}`));
  }

  return stream
    .pipe(
      htmlmin({
        collapseWhitespace: true,
        conservativeCollapse: true,
        decodeEntities: false,
        keepClosingSlash: true,
        removeComments: true,
        removeRedundantAttributes: true,
        sortAttributes: true,
        sortClassName: true,
        minifyCSS: true,
        minifyJS: true,
        minifyHTML: true,
      })
    )
    .pipe(dest(DIST_DIR))
    .pipe(
      exec((file) => `go run -C ${JSON.stringify(PREMAILER_DIR)} . ${JSON.stringify(file.path)}`)
    )
    .pipe(exec.reporter());
}

export default series(clean, minifyCSS, minifyImages, minifyHTML);
