import { defineCommand } from 'clerc';
import { Logger } from '../../services/LoggerService.ts';
import { UserConfigFile } from '../../services/ConfigService.ts';
import { promises as fs } from 'fs';
import { dirname } from 'node:path';

export const InitCommand = defineCommand(
  {
    name: 'init',
    description: 'Initialize wiki configuration',
    flags: {},
    alias: [],
    parameters: [],
  },
  async (ctx) => {
    Logger.debug('InitCmd called');

    try {
      const configService = ctx.store.config;

      if (!configService) {
        Logger.error('[ERROR] Config service not available');
        return;
      }

      // Create config directory if it doesn't exist
      await fs.mkdir(dirname(UserConfigFile), { recursive: true });

      // Write default config
      const defaultConfig = configService.store;
      await configService.write(defaultConfig);

      Logger.info(`Wiki initialized at ${UserConfigFile}`);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.toString() : String(error);
      Logger.error(`[ERROR] Failed to initialize wiki: ${errorMessage}`);
    }
  }
);
