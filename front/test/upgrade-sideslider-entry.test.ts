// @vitest-environment node
import { existsSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it } from 'vitest';

const frontRoot = resolve(__dirname, '..');

const readSource = (relativePath: string) => readFileSync(resolve(frontRoot, relativePath), 'utf-8');

describe('upgrade sideslider entry wiring', () => {
  it('uses the new upgrade sideslider component in node agent list', () => {
    const source = readSource('src/pages/node/agent/list.vue');

    expect(source).toContain('<upgrade-sideslider');
    expect(source).toContain("import UpgradeSideslider from './upgrade-sideslider.vue';");
    expect(source).not.toContain('<upgrade-form-sideslider');
    expect(source).not.toContain('<upgrade-preview-sideslider');
  });

  it('uses the new upgrade sideslider component in workarea more action', () => {
    const source = readSource('src/pages/topo/workarea-detail/components/more-action.vue');

    expect(source).toContain('<upgrade-sideslider');
    expect(source).toContain("import UpgradeSideslider from '@/pages/node/agent/upgrade-sideslider.vue';");
    expect(source).not.toContain('<upgrade-form-sideslider');
    expect(source).not.toContain('<upgrade-preview-sideslider');
  });

  it('defines a dedicated upgrade sideslider component', () => {
    const componentPath = resolve(frontRoot, 'src/pages/node/agent/upgrade-sideslider.vue');

    expect(existsSync(componentPath)).toBe(true);
    expect(readFileSync(componentPath, 'utf-8')).toContain('<Sideslider');
  });
});
