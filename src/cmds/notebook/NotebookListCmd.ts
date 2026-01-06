import { defineCommand } from 'clerc';
import { Logger } from '../../services/LoggerService';

const Log = Logger.child({ namespace: 'NotebookListCmd' });

export const NotebookListCommand = defineCommand(
  {
    name: 'notebook list',
    description: 'Manage wiki notebooks',
    flags: {},
    alias: ['nb list'],
    parameters: [],
  },
  async (ctx) => {
    Log.debug('Execute');
    const notebooks = await ctx.store.notebooKService?.list();

    if (!notebooks) {
      Log.error('No notebook');
      return;
    }

    if (notebooks.length === 0) {
      Log.debug('NotebookListCmd: found %d notebooks', notebooks.length);
      // eslint-disable-next-line no-console
      console.log('No notebooks found.');
      return;
    }

    for (const notebook of notebooks) {
      // eslint-disable-next-line no-console
      console.log(`- ${notebook.path} (${notebook.config.name})`);
    }
  }
);
