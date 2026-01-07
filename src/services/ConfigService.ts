import type { Config as ConfigShape } from 'bunfig';
import nconf from 'nconf';
import { type } from 'arktype';
import { join } from 'node:path';
import envPaths from 'env-paths';
import { Logger } from './LoggerService';
import { prettifyArktypeErrors } from '../core/schema';

const Log = Logger.child({ namespace: 'ConfigService' });

const Paths = envPaths('opennotes', { suffix: '' });

/**
 * Config File found in the user's config directory
 * We consider this to be the global config
 */
export const GlobalConfigFile = join(Paths.config, 'config.json');
/**
 * Config File
 *
 * These look like:
 *
 * .opennotes.json (path: ./SomeNotebook)
 * SomeNotebook/
 *   SomeNote.md
 *   AnotherNote.md
 *   AFolder/
 *     NestedNote.md
 *     etc.md
 *
 * or
 *
 * AnotherNotebook/
 *   .opennotes.json (path: ./)
 *   Notes.md
 */
export const NotebookConfigFile = '.opennotes.json';

export const ConfigSchema = type({
  /**
   * Notebook paths are directories where there is a NotebookConfigFile
   */
  notebooks: 'string[]',
  /**
   * Current notebook path
   *
   * This can be provided from either (first one found wins):
   * - OPENNOTES_NOTEBOOK_PATH env variable
   * - --notebookPath CLI flag
   * - or stored in global config.
   */
  notebookPath: 'string?',
});

export type Config = typeof ConfigSchema.infer;

type ConfigWriter = (config: Config) => Promise<void>;

export type ConfigService = {
  store: Config;
  write: ConfigWriter;
};

const options: ConfigShape<Config> = {
  name: 'opennotes',
  cwd: './',
  defaultConfig: {
    notebooks: [join(Paths.config, 'notebooks')],
  },
};
export async function createConfigService(): Promise<ConfigService> {
  const config = nconf.env({
    separator: '__',
    parseValues: true,
    lowerCase: true,
    match: /^opennotes_/i,
    transform: (obj: { key: string; value: string }) => {
      obj.key = obj.key.replace(/^opennotes_/, '');
      return obj;
    },
  });
  Log.debug('Config after env: %o', { config: config.get() });

  nconf.file({ file: GlobalConfigFile });

  Log.debug('Config after file: [%s] %o', GlobalConfigFile, { config: nconf.get() });

  nconf.defaults(options.defaultConfig);

  Log.debug('Config loaded from defaults: %o', { config: nconf.get() });
  const value = nconf.get();

  Log.debug('Loaded config: %o', { value });
  const store = ConfigSchema(value);

  if (store instanceof type.errors) {
    Log.error('Invalid config file at %s: \n %s', GlobalConfigFile, prettifyArktypeErrors(store));
    throw new Error(`Invalid config file at ${GlobalConfigFile}`);
  }

  async function write(config: Config): Promise<void> {
    return new Promise((resolve, reject) => {
      Log.debug('Writing: %o', { config });
      nconf.save((err: Error) => {
        if (err) {
          Log.error('Failed to write config: %o', err);
          return reject(err);
        }

        Log.info('Config written to %s', GlobalConfigFile);
        resolve();
      });
    });
  }

  Log.debug('Ready %o', store);

  return {
    store,
    write,
  };
}
