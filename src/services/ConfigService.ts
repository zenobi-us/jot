import type { Config as ConfigShape } from 'bunfig';
import { loadConfig } from 'bunfig';
import { type } from 'arktype';
import { promises as fs } from 'fs';
import { join, dirname } from 'node:path';
import envPaths from 'env-paths';
import { Logger } from './LoggerService';

const Log = Logger.child({ namespace: 'ConfigService' });

const Paths = envPaths('wiki', { suffix: '' });

export const UserConfigFile = join(Paths.config, 'config.json');

export const ConfigSchema = type({
  notebooks: 'string[]',
  notebookPath: 'string?',
  configFilePath: 'string',
});

export type Config = typeof ConfigSchema.infer;

type ConfigWriter = (config: Config) => Promise<void>;

export type ConfigService = {
  store: Config;
  write: ConfigWriter;
};

const options: ConfigShape<Config> = {
  name: 'opentask',
  cwd: './',
  defaultConfig: {
    notebooks: [join(Paths.config, 'notebooks')],
    configFilePath: 'wiki/config.json',
  },
};

export async function createConfigService(): Promise<ConfigService> {
  Log.debug('Loading');
  const store = await loadConfig(options);
  Log.debug('Loadeded %o', { store });

  async function write(config: Config): Promise<void> {
    Log.debug('Writing: %o', { config });
    await fs.mkdir(dirname(UserConfigFile), { recursive: true });
    await fs.writeFile(UserConfigFile, JSON.stringify(config, null, 2));
    Log.debug('Written');
  }

  Log.debug('Ready');

  return {
    store,
    write,
  };
}
