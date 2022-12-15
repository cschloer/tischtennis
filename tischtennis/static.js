"use strict";
const path = require("path");
const StaticFileHandler = require("serverless-aws-static-file-handler");
const clientFilesPath = path.join(__dirname, "./static/");
const fileHandler = new StaticFileHandler(clientFilesPath);

module.exports.staticHandler = async (event, context) => {
  if (!event.path.startsWith("/static/")) {
    throw new Error(`[404] Invalid filepath for this resource`);
  }
  return fileHandler.get(event, context);
};
