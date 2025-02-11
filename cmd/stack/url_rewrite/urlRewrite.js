const SUPPORTED_FORMATS = ["auto", "avif", "jpeg", "jpg", "webp", "png", "svg"];

const OPERATIONS = {
  format: "f",
  width: "w",
  height: "h",
  quality: "q",
};

async function handler(event) {
  const request = event.request;
  const originalImagePath = request.uri;
  const queryString = request.querystring;
  const normalizedOperations = {};

  if (queryString) {
    var queryStringKeys = Object.keys(queryString);

    for (var i = 0; i < queryStringKeys.length; i++) {
      var operation = queryStringKeys[i].toLowerCase();
      var value =
        queryString[operation] && queryString[operation].value.toLowerCase();

      if (!value) continue;

      switch (operation) {
        case OPERATIONS.format:
          if (SUPPORTED_FORMATS.indexOf(value) === -1) continue;
          break;
        case OPERATIONS.width:
        case OPERATIONS.height:
        case OPERATIONS.quality:
          var numValue = parseInt(value, 10);
          if (isNaN(numValue) || numValue <= 0) continue;
          if (operation === OPERATIONS.quality && numValue > 100)
            numValue = 100;
          value = numValue.toString();
          break;
      }

      normalizedOperations[operation] = value;
    }

    if (Object.keys(normalizedOperations).length > 0) {
      var normalizedOperationsArray = [];

      if (normalizedOperations[OPERATIONS.format]) {
        normalizedOperationsArray.push(
          OPERATIONS.format + "=" + normalizedOperations[OPERATIONS.format]
        );
      }
      if (normalizedOperations[OPERATIONS.width]) {
        normalizedOperationsArray.push(
          OPERATIONS.width + "=" + normalizedOperations[OPERATIONS.width]
        );
      }
      if (normalizedOperations[OPERATIONS.height]) {
        normalizedOperationsArray.push(
          OPERATIONS.height + "=" + normalizedOperations[OPERATIONS.height]
        );
      }
      if (normalizedOperations[OPERATIONS.quality]) {
        normalizedOperationsArray.push(
          OPERATIONS.quality + "=" + normalizedOperations[OPERATIONS.quality]
        );
      }

      request.uri =
        originalImagePath + "?" + normalizedOperationsArray.join("&");
    }
  }
  return request;
}
