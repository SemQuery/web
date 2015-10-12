var gulp = require('gulp'),
    concat = require('gulp-concat'),
    rename = require('gulp-rename'),
    uglify = require('gulp-uglify'),
    minify_css = require('gulp-minify-css'),
    sass = require('gulp-sass'),
    svgmin = require('gulp-svgmin');

var paths = {
    js: ['assets/js/**/*.js'],
    css: ['assets/css/**/*.css'],
    scss: ['assets/css/**/*.scss'],
    images: [
        'assets/images/semquery_logo_shadow.png',
        'assets/images/logo.png'
    ],
    svgs: [
        'assets/images/tiled_carets.svg',
        'assets/images/code_pages.svg',
        'assets/images/code_pages_query.svg',
        'assets/images/logo_v2.svg',
        'assets/images/logo_v2_white.svg',
        'assets/images/logo_v2_small.svg'
    ]
};

var dests = {
    js: 'public/js',
    css: 'public/css',
    images: 'public/images'
};

gulp.task('js', function() {
    return gulp.src(paths.js)
        .pipe(concat('all.js'))
        .pipe(uglify())
        .pipe(gulp.dest(dests.js));
});

gulp.task('css', function() {
    return gulp.src(paths.scss)
        .pipe(sass({outputStyle: 'compressed'}))
        .pipe(concat('all.css'))
        .pipe(gulp.dest(dests.css));
});

gulp.task('img', function() {
    return gulp.src(paths.images)
        .pipe(gulp.dest(dests.images));
});

gulp.task('svg', function() {
    return gulp.src(paths.svgs)
        .pipe(svgmin())
        .pipe(gulp.dest(dests.images));
});

gulp.task('images', ['img', 'svg'], function() {});

gulp.task('default', ['js', 'css', 'images'], function(){});
