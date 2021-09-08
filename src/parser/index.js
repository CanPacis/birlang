const { BirParser } = require("./parser");

let parser = new BirParser();

if (process.argv[2]) {
  try {
    let result = parser.parse(process.argv[2]);
    console.log(JSON.stringify({ error: false, content: result }));
  } catch (error) {
    let lines = error.message.split("\n");
    let line = lines[0].split("line")[1].split(" ")[1].trim();
    let col = lines[0]
      .split("col")[1]
      .split("")
      .reverse()
      .join("")
      .substr(1)
      .split("")
      .reverse()
      .join("")
      .trim();
    let message = lines[4].split(". Instead")[0];
    console.log(JSON.stringify({ error: true, content: { message, position: { line: parseInt(line), col: parseInt(col) } } }));
  }
} else {
  console.log(JSON.stringify({ error: false, content: {imports:[], program: []} }));
}
