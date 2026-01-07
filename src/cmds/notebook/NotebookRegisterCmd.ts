import { defineCommand } from '@clerc/core';
import { Logger } from '../../services/LoggerService.ts';

const Log = Logger.child({ namespace: 'NotebookRegisterCommand' });

export const NotebookRegisterCommand = defineCommand(
  {
    name: 'notebook register',
    description: 'Register an existing notebook folder globally',
    alias: ['nb register'],
    parameters: ['<path>'],
  },
  async (ctx) => {
    const notebookPath = ctx.parameters.path;

    Log.debug('Execute: register notebook at %s', notebookPath);

    const notebook = await ctx.store.notebooKService?.open(notebookPath);

    if (!notebook) {
      Log.error('Failed to load notebook at %s', notebookPath);
      return;
    }

    await notebook.saveConfig({ register: true });

    Log.info('Registered notebook: %o', { notebook });
  }
);
