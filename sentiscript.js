const sentiment = require('sentiment');
const fs = require('fs');

// const r1 = sentiment('Cats are stupid.');
// console.dir(r1);
//
// const r2 = sentiment('Cats are totally amazing!');
// console.dir(r2);
//
// process.argv.forEach(function (val, index, array) {
//   console.log(index + ': ' + val);
// });

//Note: the below example reads the entire file in at once. For a more efficient
// solution, see: http://stackoverflow.com/questions/6156501/read-a-file-one-line-at-a-time-in-node-js
fs.readFile(process.argv[2], 'utf8', function(err, text) {
    if (err) throw err;
    console.log(sentiment(text).score);
});
