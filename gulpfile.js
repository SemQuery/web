var gulp = require('gulp'),
    concat = require('gulp-concat'),
    rename = require('gulp-rename'),
    uglify = require('gulp-uglify');

var paths = {
    js: ['assets/js/**/*.js']
};

var dests = {
    js: 'public/js'
};

gulp.task('js', function() {
    return gulp.src(paths.js)
        .pipe(concat('concat.js'))
        .pipe(gulp.dest(dests.js))
        .pipe(rename('uglify.js'))
        .pipe(uglify())
        .pipe(gulp.dest(dests.js));
});

gulp.task('default', ['js'], function(){});
