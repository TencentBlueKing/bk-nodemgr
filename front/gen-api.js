const fs = require('fs');
const path = require('path');
const protobuf = require('protobufjs');

const protoDir = '../../proto'; // Directory containing proto files

// Function to parse proto files and generate output
async function generateFiles() {
  const protoFiles = fs.readdirSync(path.resolve(__dirname, protoDir)).filter(file => file.endsWith('.proto'));

  for (const file of protoFiles) {
    const filePath = path.join(protoDir, file);
    // Read the file content and remove import statements
    let fileContent = fs.readFileSync(filePath, 'utf-8');
    fileContent = fileContent.split('\n').filter(line => !line.trim().startsWith('import'))
      .join('\n');

    const { root } = protobuf.parse(fileContent, { alternateCommentMode: true });

    if (!root || !root.nestedArray || root.nestedArray.length === 0) {
      console.warn(`No data found in proto file: ${file}`);
      continue;
    }

    const fileName = path.basename(file, '.proto');

    // Generate TypeScript definitions
    const typeDefinitions = generateTypeDefinitions(root);
    const typeFileName = path.join('./src/@types', `${fileName}.d.ts`);
    fs.writeFileSync(typeFileName, typeDefinitions);

    // Generate API interface definitions
    const apiDefinitions = generateApiDefinitions(root, fileName);
    const apiFileName = path.join('./src/api/modules', `${fileName}.ts`);
    fs.writeFileSync(apiFileName, apiDefinitions);
  }
}

// Function to generate TypeScript type definitions
function generateTypeDefinitions(root) {
  let output = '// gen-api.js 自动生成，请勿手动修改\n';
  root.nestedArray.forEach((namespace) => {
    namespace.nestedArray.forEach((namespace) => {
      namespace.nestedArray.forEach((namespace) => {
        output += `// ${namespace.name}\n`;
        namespace.nestedArray.forEach((type) => {
          if (type instanceof protobuf.Type) {
            output += `export interface ${type.name} {\n`;
            type.fieldsArray.forEach((field) => {
              const commentList = field.comment?.split('\n');
              output += commentList?.length ? `${commentList.map(comment => `  // ${comment}\n`).join('')}` : '';
              output += `  ${field.name}: ${mapProtoTypeToTs(field)};\n`;
            });
            output += '}\n\n';
          }
        });
      });
    });
  });
  return output;
}

const allHttpMethods = ['get', 'post', 'put', 'delete', 'patch'];
// Function to generate API interface definitions
function generateApiDefinitions(root, fileName) {
  let output = `// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

`;

  let tmpOutput = '';
  const types = [];
  root.nestedArray.forEach((namespace) => {
    namespace.nestedArray.forEach((namespace) => {
      namespace.nestedArray.forEach((namespace) => {
        namespace.nestedArray.forEach((service) => {
          if (service instanceof protobuf.Service) {
            tmpOutput += `export const ${service.name}Service = {\n`;
            service.methodsArray.forEach((method) => {
              const trpcOpts = method.parsedOptions?.find(opt => opt['(trpc.api.http)']);
              let httpMethod = '';
              let path = '';
              if (trpcOpts) {
                httpMethod = allHttpMethods.find(m => trpcOpts['(trpc.api.http)']?.[m]);
                path = trpcOpts['(trpc.api.http)']?.[httpMethod];
              }
              if (!types.includes(method.requestType)) {
                types.push(method.requestType);
              }
              if (!types.includes(method.responseType)) {
                types.push(method.responseType);
              }
              const commentList = method.comment?.split('\n');
              tmpOutput += commentList?.length ? `${commentList.map(comment => `  // ${comment}\n`).join('')}` : '';
              tmpOutput += `  ${method.name}: async <Request = ${method.requestType}, ResponseData = ${method.responseType}['datas']>(params?: Request, config?: Config) => await fetch.${httpMethod}<Request, ResponseData>('${path}')(params, config),\n`;
            });
            tmpOutput += '};\n\n';
          }
        });
      });
    });
  });

  output += `import type { ${types.join(', ')} } from '@/@types/${fileName}';\n`;

  output += `import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: \`\${import.meta.env.BK_API_PREFIX}\`,
});\n\n`;

  output += tmpOutput;

  return output;
}

// Helper function to map proto types to TypeScript types
function mapProtoTypeToTs(field) {
  if (field.map) {
    return `Record<${field.keyType}, ${mapBasicProtoTypeToTs(field.type)}>`;
  }
  if (field.repeated) {
    return `${mapBasicProtoTypeToTs(field.type)}[]`;
  }
  const { type } = field;
  return mapBasicProtoTypeToTs(type);
}

function mapBasicProtoTypeToTs(type) {
  switch (type) {
    case 'string':
      return 'string';
    case 'int32':
    case 'int64':
    case 'uint32':
    case 'uint64':
    case 'float':
    case 'double':
      return 'number';
    case 'bool':
      return 'boolean';
    case 'google.protobuf.Struct':
      return 'Record<string, any>';
    case 'google.protobuf.Timestamp':
      return 'Date';
    default:
      return type; // For complex types, return as is
  }
}

generateFiles().catch(console.error);
