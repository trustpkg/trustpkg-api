import fs from "node:fs";

const filePath = process.argv[2] ?? "postman/trust-pkg-api.json";
const baseUrl = process.argv[3] ?? "http://localhost:3001";
const collection = JSON.parse(fs.readFileSync(filePath, "utf8"));

collection.variable ??= [];
const baseUrlVariable = collection.variable.find((variable) => variable.key === "baseUrl");
if (baseUrlVariable) {
  baseUrlVariable.value = baseUrl;
} else {
  collection.variable.push({ key: "baseUrl", value: baseUrl });
}

fs.writeFileSync(filePath, `${JSON.stringify(collection, null, 2)}\n`);