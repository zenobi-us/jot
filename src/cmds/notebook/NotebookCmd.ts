import { defineCommand } from 'clerc';
import { Logger } from '../../services/LoggerService';
import { requireNotebookMiddleware } from '../../middleware/requireNotebookMiddleware';

const Log = Logger.child({ namespace: 'NotebookCmd' });

export const NotebookCommand = defineCommand(
  {
    name: 'notebook',
    description: 'Manage wiki notebooks',
    flags: {},
    alias: ['nb'],
    parameters: [],
  },
  async (ctx) => {
    Log.debug('Execute');

    const notebook = await requireNotebookMiddleware({
      notebookService: ctx.store.notebooKService,
      path: ctx.flags.notebook,
    });

    if (!notebook) {
      return;
    }

    await ctx.store.notebooKService?.info({ notebook });
  }
);
