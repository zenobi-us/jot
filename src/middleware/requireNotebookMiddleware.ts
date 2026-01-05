import { dedent } from '../core/strings';
import { RenderMarkdownTui } from '../services/Display';
import type { NotebookService } from '../services/NotebookService';

export async function requireNotebookMiddleware(args: {
  path?: string;
  notebookService?: NotebookService;
}) {
  let notebook = args?.path
    ? await args.notebookService?.open(args.path)
    : await args.notebookService?.infer();

  if (!notebook) {
    const message = await RenderMarkdownTui(
      dedent(`

        # No Notebook Yet
        
        If you want to start using notebooks to manage your wiki, you first need to create a notebook.
        
        You can create a new notebook by running the following command:

        \`\`\`bash
        wiki notebook create [path]
        \`\`\`
    `)
    );
    // eslint-disable-next-line no-console
    console.error(message);
    return null;
  }

  return notebook;
}
