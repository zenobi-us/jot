import { defineCommand } from 'clerc';
import { Logger } from '../../services/LoggerService';
import { slugify } from '../../core/strings';
import path from 'node:path';

const Log = Logger.child({ namespace: 'NotebookCreateCmd' });
/**
 * Command to create a new notebook in the wiki system.
 *
 *
 */
export const NotebookCreateCommand = defineCommand(
  {
    name: 'notebook create',
    description: 'Manage wiki notebooks',
    flags: {
      name: {
        description: 'Name of the notebook to create',
        type: String,
        required: false,
      },
    },
    alias: ['nb create'],
    parameters: ['[path]'],
  },
  async (ctx) => {
    const notebookPath = ctx.parameters.path || process.cwd();
    const notebookName = ctx.flags.name || slugify(path.basename(notebookPath));

    Log.debug('Execute: %s, %s', notebookName, notebookPath);

    const notebook = await ctx.store.notebooKService?.create({
      name: notebookName,
      path: notebookPath,
    });

    Log.info('Created notebook: %o', { notebook });
  }
);
