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
    description: 'Create a new notebook or register an existing folder',
    flags: {
      name: {
        description: 'Name of the notebook',
        type: String,
        required: false,
      },
      register: {
        description: 'Register this notebook globally',
        type: Boolean,
        required: false,
      },
    },
    alias: ['nb create'],
    parameters: ['[path]'],
  },
  async (ctx) => {
    const notebookPath = ctx.parameters.path || process.cwd();
    const notebookName = ctx.flags.name || slugify(path.basename(notebookPath));
    const register = ctx.flags.register || false;

    Log.debug('Execute: %s, %s, register=%s', notebookName, notebookPath, register);

    const notebook = await ctx.store.notebooKService?.create({
      name: notebookName,
      path: notebookPath,
      register,
    });

    Log.info('Created notebook: %o', { notebook });
  }
);
