import type { Config as ConfigShape } from 'bunfig';
import { loadConfig } from 'bunfig';
import { type } from 'arktype';
import { promises as fs } from 'fs';
import { join, dirname } from 'node:path';
import envPaths from 'env-paths';

const Paths = envPaths('wiki', { suffix: '' });

export const UserConfigFile = join(Paths.config, 'config.json');

export const ConfigSchema = type({
  notebooks: 'string[]',
  notebookPath: 'string?',
  configFilePath: 'string',
});

export type Config = typeof ConfigSchema.infer;

export type ConfigService = {
  store: Config;
  write: (config: Config) => Promise<void>;
};

const options: ConfigShape<Config> = {
  name: 'opentask',
  cwd: './',
  defaultConfig: {
    notebooks: [join(Paths.config, 'notebooks')],
    configFilePath: 'wiki/config.json',
  },
};

export async function createConfigService(args: { directory: string }): Promise<ConfigService> {
  const store = await loadConfig(options);

  async function write(config: Config): Promise<void> {
    await fs.mkdir(dirname(UserConfigFile), { recursive: true });
    await fs.writeFile(UserConfigFile, JSON.stringify(config, null, 2));
  }

  return {
    store,
    write,
  };
}
