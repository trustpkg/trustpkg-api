import fs from "node:fs";

const filePath = process.argv[2] ?? process.env.OPENAPI_FILE ?? "docs/openapi.json";
const document = JSON.parse(fs.readFileSync(filePath, "utf8"));
const npmPackagePattern = "^@?[a-z0-9][a-z0-9._-]*(/[a-z0-9][a-z0-9._-]*)?$";

const searchOperation = document.paths?.["/packages/search"]?.get;
const queryParameter = searchOperation?.parameters?.find(
  (parameter) => parameter.name === "query",
);
if (queryParameter?.schema) {
  queryParameter.schema.pattern = npmPackagePattern;
}

const searchDocument = document.components?.schemas?.["packagesApi.SearchDocument"];
const nameSchema = searchDocument?.properties?.name;
if (nameSchema) {
  nameSchema.pattern = npmPackagePattern;
}

const vulnerabilityValues = searchDocument?.properties?.vulnerabilities?.additionalProperties;
if (vulnerabilityValues) {
  vulnerabilityValues.minimum = 1;
}

fs.writeFileSync(filePath, `${JSON.stringify(document, null, 2)}\n`);