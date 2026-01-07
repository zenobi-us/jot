import { defineCommand } from 'clerc';
import { Logger } from '../../services/LoggerService';
import { requireNotebookMiddleware } from '../../middleware/requireNotebookMiddleware';
import { TuiTemplates as NotesTuiTemplates } from '../../services/NoteService';

export const NotesListCommand = defineCommand(
  {
    name: 'notes list',
    description: 'List all notes in the project',
    flags: {},
    alias: [],
    parameters: [],
  },
  async (ctx) => {
    const notebook = await requireNotebookMiddleware({
      notebookService: ctx.store.notebooKService,
      path: ctx.flags.notebook,
    });

    if (!notebook) {
      return;
    }

    Logger.debug('NotesListCmd %s', notebook.config.path);

    const configService = ctx.store.config;
    const dbService = ctx.store.dbService;

    if (!configService || !dbService) {
      // eslint-disable-next-line no-console
      console.error('Failed to load config or dbService');
      return;
    }

    const results = await notebook.notes.searchNotes();

    // eslint-disable-next-line no-console
    console.log(
      await NotesTuiTemplates.NoteList({
        notes: results,
      })
    );
  }
);
