import { readdirSync, readFileSync, statSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join, resolve } from 'node:path';

import { expect, it } from 'vitest';

import { productRepo } from './shared';

const selfPath = fileURLToPath(import.meta.url);

function collect(dir: string, out: string[]): void {
  for (const entry of readdirSync(dir)) {
    if (entry === 'node_modules' || entry === '.next') continue;
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      collect(full, out);
      continue;
    }
    if (resolve(full) === selfPath) continue;
    if (entry.endsWith('.mdx') || entry.endsWith('.ts') || entry.endsWith('.tsx')) {
      out.push(readFileSync(full, 'utf8'));
    }
  }
}

it('docs do not link the upstream donate page or Telegram channel', () => {
  const docsRoot = resolve(process.cwd());
  const files: string[] = [];
  collect(join(docsRoot, 'content'), files);
  collect(join(docsRoot, 'lib'), files);
  collect(join(docsRoot, 'app'), files);
  const blob = files.join('\n');
  expect(blob).not.toContain('donate.sanaei.dev');
  expect(blob).not.toContain('t.me/XrayUI');
  expect(blob).not.toContain('donateUrl');
  expect(blob).not.toContain('telegramChannelUrl');
});

it('product repository is this fork', () => {
  expect(productRepo).toEqual({ user: 'yobo2u', repo: '3x-ui', branch: 'main' });
});
