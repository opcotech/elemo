import { rmSync, mkdirSync } from 'node:fs';
import { dirname } from 'node:path';

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

const BASE_PATH = dirname('.');
const BUILD_DIR = `${BASE_PATH}/build`;
const SRC_DIR = `${BASE_PATH}/src`;
const DIST_DIR = options.out;

const IMAGE_FILES = `${SRC_DIR}/**/*.(png|jpg|jpeg)`;
const CSS_FILES = `${SRC_DIR}/**/*.css`;
const HTML_FILES = [`${SRC_DIR}/**/*.html`, `!${SRC_DIR}/includes/*.html`];

const BUCKET_STATIC_PATH = 'email-assets/';

function clean(cb) {
  rmSync(BUILD_DIR, { force: true, recursive: true });
  mkdirSync(BUILD_DIR);

  cb();
}

function minifyImages(cb) {
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

  stream.pipe(dest(BUILD_DIR));

  cb();
}

function minifyCSS(cb) {
  src(CSS_FILES)
    .pipe(purgecss({ content: [HTML_FILES] }))
    .pipe(dest(BUILD_DIR));

  cb();
}

function minifyHTML(cb) {
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

  stream
    .pipe(
      htmlmin({
        collapseWhitespace: true,
        conservativeCollapse: true,
        decodeEntities: true,
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
    .pipe(exec((file) => `cd ../../tools/pre-mailer/; go run main.go ${file.path}`));

  cb();
}

export default series(clean, minifyCSS, minifyImages, minifyHTML);
