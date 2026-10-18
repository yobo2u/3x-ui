import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it } from 'vitest';

const src = resolve(process.cwd(), 'src');

function source(path: string): string {
  return readFileSync(resolve(src, path), 'utf8');
}

describe('paid sponsor and donate surfaces', () => {
  const files = [
    'pages/login/LoginPage.tsx',
    'pages/index/IndexPage.tsx',
    'layouts/AppSidebar.tsx',
    'routes.tsx',
    'components/command-palette/CommandPalette.tsx',
  ];

  it('panel pages no longer mount a sponsor slot or donate link', () => {
    const blob = files.map(source).join('\n');
    expect(blob).not.toContain('SponsorSlot');
    expect(blob).not.toContain('donate.sanaei.dev');
    expect(blob).not.toContain("'/sponsors'");
    expect(blob).not.toContain('"/sponsors"');
    expect(blob).not.toContain('menu.sponsors');
    expect(blob).not.toContain('menu.donate');
  });
});
