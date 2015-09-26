var gulp = require('gulp'),
    concat = require('gulp-concat'),
    rename = require('gulp-rename'),
    uglify = require('gulp-uglify'),
    minify_css = require('gulp-minify-css');

var paths = {
    js: ['assets/js/**/*.js'],
    css: ['assets/css/**/*.css'],
    images: [
        'assets/images/semquery_logo_shadow.png',
        'assets/images/tiled_carets.svg',
        'assets/images/logo.png'
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
    return gulp.src(paths.css)
        .pipe(concat('all.css'))
        .pipe(minify_css())
        .pipe(gulp.dest(dests.css));
});

gulp.task('images', function() {
    return gulp.src(paths.images)
        .pipe(gulp.dest(dests.images));
});

gulp.task('default', ['js', 'css', 'images'], function(){});
