function includesStr(line, index) {
    let startIndex;
    if (line[index] != '"') {
        // Find the start index of the substring starting at the given index    
        let substring = line.substring(0, index);
        // Reverse the substring
        let lookBack = substring.split("").reverse().join("");
        startIndex = lookBack.length - (lookBack.indexOf('"') + 1);
        if (startIndex === -1) {
            // No substring starting at the given index
            return false, 0, 0;
        }
    
        // Find the end index of the substring
        let endIndex = line.indexOf('"', index + 1);
        if (endIndex === -1) {
            // No end quote found, so the substring is not properly closed
            return false, 0, 0;
        }

      quoteCount = 0
      for (let i = 0; i < line.length; i++){
        if (line[i] == '"') {
          quoteCount++
        }
        if (i == endIndex && quoteCount % 2 == 0) {
          console.log(startIndex,endIndex)
          return true, startIndex, endIndex;
        }
      }
      return false ,0,0
    } else {
        return true
    }
}
console.log(
  includesStr(
    'case "(" , ")" , "{" , "}" , "if" , "else" , "elif" , "case" , "switch":',
    7
  )
);
