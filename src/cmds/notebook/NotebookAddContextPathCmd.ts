import { defineCommand } from 'clerc';
import { Logger } from '../../services/LoggerService';

export const NotebookAddContextPathCommand = defineCommand(
  {
    name: 'notebook add-context-path',
    description: 'Add context path to a notebook.',
    flags: {
      notebookPath: {
        type: String,
        description: 'Path to the notebook',
        required: true,
      },
    },
    alias: ['nb add-context'],
    parameters: ['[path]'],
  },
  () => {
    Logger.debug('Notebook add-context-path command executed');
  }
);
