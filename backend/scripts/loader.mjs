// Node loader: lets the exporter import the frontend's logic files unchanged ("~/" means the project root).
import { pathToFileURL } from 'node:url';
import { existsSync } from 'node:fs';
import path from 'node:path';
const ROOT = path.resolve(import.meta.dirname, '..', '..');
export async function resolve(spec, ctx, next) {
  if (spec.startsWith('~/composables/')) return { url: 'data:text/javascript,export const useAccounts=()=>({});export const useAuth=()=>({});', shortCircuit: true };
  if (spec.startsWith('~/')) {
    let p = path.join(ROOT, spec.slice(2));
    for (const ext of ['.js', '.ts', '.mjs']) { if (!existsSync(p) && existsSync(p + ext)) { p += ext; break; } }
    return { url: pathToFileURL(p).href, shortCircuit: true };
  }
  return next(spec, ctx);
}
