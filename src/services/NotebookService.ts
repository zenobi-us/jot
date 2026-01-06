import type { Config } from './ConfigService';
import { promises as fs } from 'fs';
import { dirname, join } from 'path';
import { type } from 'arktype';
import { dedent } from '../core/strings';
import { Logger } from './LoggerService';
import { TuiRender } from './Display';

const Log = Logger.child({ namespace: 'NotebookService' });

const NotebookGroupSchema = type({
  name: 'string',
  description: 'string?',
  globs: 'string[]',
  metadata: type({ '[string]': 'string | number | boolean' }),
});

export type NotebookGroup = typeof NotebookGroupSchema.infer;

const NotebookConfigSchema = type({
  name: 'string',
  contexts: 'string[]?',
  templates: type({ '[string]': 'string' }).optional(),
  groups: NotebookGroupSchema.array().optional(),
});

type NotebookConfig = typeof NotebookConfigSchema.infer;

export interface INotebook {
  path: string;
  config: NotebookConfig;
  saveConfig(): Promise<void>;
  addContext(_contextPath?: string): Promise<void>;
  loadTemplate(_name: string): Promise<string | null>;
}

export const TuiTemplates = {
  NotebookCreated: (ctx: { name: string; path: string }) =>
    TuiRender(
      dedent(`
    # Notebook Created

    Your new notebook has been successfully created!

    - **Name**: {{name}}
    - **Path**: {{path}}

    You can start adding notes to your notebook right away.
  `),
      ctx
    ),
  ContextAlreadyExists: (ctx: { contextPath: string; notebookPath: string }) =>
    TuiRender(
      dedent(`
    # Context Already Exists

    The context path is already associated with this notebook.

    - **Context**: {{contextPath}}
    - **Notebook**: {{notebookPath}}

    No changes were made.
  `),
      ctx
    ),
  ContextAdded: (ctx: { contextPath: string; notebookPath: string }) =>
    TuiRender(
      dedent(`
    # Context Added

    The context path has been successfully added to your notebook.

    - **Context**: {{contextPath}}
    - **Notebook**: {{notebookPath}}

    This notebook will now be available when working in that directory.
  `),
      ctx
    ),
  TemplateLoadError: (ctx: { templatePath: string; error: string }) =>
    TuiRender(
      dedent(`
    # Template Load Error

    Failed to load a template for your notebook. This may cause some features to be unavailable.

    - **Template Path**: {{templatePath}}
    - **Error**: {{error}}

    You may need to check the template file and try again.
  `),
      ctx
    ),
  NotebookInfo: (notebook: INotebook) =>
    TuiRender(
      dedent(`
    # Notebook Information

    - **Name**: {{notebook.config.name}}
    - **Path**: {{notebook.path}}

    ## Contexts

    {% for context in notebook.config.contexts %}
      - {{ context }}
    {% empty %}
      - _No contexts defined_
    {% endfor %}

    ## Groups

    {% for group in notebook.config.groups %}
      - **Name**: {{group.name}}
        - Description: {{group.description}}
        - Globs:
        {% for glob in group.globs %}
          - {{ glob }}
        {% empty %}
          - _No globs defined_
        {% endfor %}
    {% endfor %}
  `),
      notebook
    ),
  CreateYourFirstNotebook: () =>
    TuiRender(
      dedent(`
    # No Notebooks Found

    It looks like you don't have any notebooks set up yet.

    To create your first notebook, use the following command:

    \`\`\`bash
    wiki notebook create --name "My First Notebook"
    \`\`\`

    This will create a new notebook in your current directory.
  `),
      {}
    ),
};

export function createNotebookService(serviceOptions: { config: Config }) {
  class Notebook implements INotebook {
    static createNotebookConfigPath(path: string) {
      return join(path, `.${serviceOptions.config.configFilePath}`);
    }

    static async isNotebookPath(path?: string) {
      if (!path) {
        return false;
      }

      const configPath = Notebook.createNotebookConfigPath(path);
      if (!(await fs.exists(configPath))) {
        return false;
      }

      return true;
    }

    static async loadConfig(path: string): Promise<NotebookConfig | null> {
      Log.debug('Notebook.loadConfig: path=%s', path);
      const configPath = Notebook.createNotebookConfigPath(path);

      try {
        const content = await fs.readFile(configPath, 'utf-8');
        const parsed = JSON.parse(content);
        const config = NotebookConfigSchema(parsed);

        if (config instanceof type.errors) {
          Log.error('NotebookService.loadNotebookConfig: INVALID_CONFIG path=%s', configPath);
          return null;
        }

        return config;
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        Log.error(
          'NotebookService.loadNotebookConfig: ERROR path=%s error=%s',
          configPath,
          errorMsg
        );
        return null;
      }
    }

    /**
     * Initialize the notebook (load config, templates, etc)
     **/
    static async load(path: string): Promise<Notebook | null> {
      const config = await Notebook.loadConfig(path);
      if (!config) {
        return null;
      }
      return new Notebook(path, config);
    }

    static async new(args: { path: string; name: string }): Promise<Notebook> {
      const config: NotebookConfig = {
        name: args.name,
        templates: {},
        groups: [
          {
            name: 'Default',
            description: 'Default group for all notes',
            globs: ['**/*.md'],
            metadata: {},
          },
        ],
        contexts: [args.path],
      };
      Log.debug('Notebook.new: path=%s name=%s', args.path, args.name);

      const notebook = new Notebook(args.path, config);
      await notebook.saveConfig();

      await TuiTemplates.NotebookCreated(args);

      return notebook;
    }

    constructor(
      public path: string,
      public config: NotebookConfig
    ) {
      //
    }

    matchContext(path: string): string | null {
      if (!this.config.contexts) return null;
      return this.config.contexts.find((context) => path.startsWith(context)) || null;
    }

    /**
     * Write a notebook config to a given path
     */
    async saveConfig() {
      const configPath = Notebook.createNotebookConfigPath(this.path);
      try {
        const content = JSON.stringify(this.config, null, 2);
        await fs.mkdir(dirname(configPath), { recursive: true });
        await fs.writeFile(configPath, content, 'utf-8');
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        Log.debug(
          'NotebookService.writeNotebookConfig: ERROR path=%s error=%s',
          configPath,
          errorMsg
        );
        throw error;
      }
    }

    /**
     * Add a context path to the notebook.
     *
     * Notebooks contexts are used to determine which notes belong to which notebooks.
     * A context is simply a path that is considered related to the notebook.
     */
    async addContext(_contextPath: string = process.cwd()): Promise<void> {
      // Check if context already exists
      if (this.config.contexts?.includes(_contextPath)) {
        await TuiTemplates.ContextAlreadyExists({
          contextPath: _contextPath,
          notebookPath: this.path,
        });

        return;
      }

      // Add the context
      this.config.contexts = [...(this.config.contexts || []), _contextPath];
      await this.saveConfig();

      await TuiTemplates.ContextAdded({
        contextPath: _contextPath,
        notebookPath: this.path,
      });
    }

    async loadTemplate(name: string): Promise<string | null> {
      if (!this.config.templates) {
        return null;
      }
      const templatePath = this.config.templates[name];
      if (!templatePath) {
        return null;
      }

      // Load templates
      // templates are listed as a mapping of template name to file path
      try {
        const template = await import(templatePath, { assert: { type: 'markdown' } }).then(
          (mod) => mod.default
        );
        return template;
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        Log.debug(
          'NotebookService.getNotebook: ERROR_LOADING_TEMPLATE path=%s error=%s',
          templatePath,
          errorMsg
        );
        await TuiTemplates.TemplateLoadError({
          templatePath,
          error: errorMsg,
        });
      }
      return null;
    }
  }

  /**
   * Get a notebook by its path
   */
  async function open(notebookPath: string): Promise<Notebook | null> {
    return Notebook.load(notebookPath);
  }

  async function create(args: { name: string; path?: string }): Promise<Notebook> {
    const notebookPath = args.path || process.cwd();
    return Notebook.new({ name: args.name, path: notebookPath });
  }

  /**
   * Discover the notebook path based on the current working directory.
   *
   * A notebook path is any folder that contains a .wiki/config.json
   *
   * Priority:
   *
   *  1. Declared Notebook Path
   *  2. Context Matching in Notebook Configs
   *  3. Ancestor Directory Search
   *
   * @param cwd Current working directory (defaults to process.cwd())
   * @returns Resolved notebook path or null if not found
   */
  async function infer(cwd: string = process.cwd()): Promise<Notebook | null> {
    Log.debug('NotebookService.infer: cwd=%s', cwd);
    const notebookPath = serviceOptions.config.notebookPath;
    // Step 1: Check environment/cli-arg variable (resolved and provided by the ConfigService)
    if (notebookPath && (await Notebook.isNotebookPath(notebookPath))) {
      const notebook = await Notebook.load(notebookPath);
      if (notebook) {
        Log.debug('Notebook.infer: USE_DECLARED_PATH %s', notebookPath);
        return notebook;
      }
    }

    for (const notebook of await list(cwd)) {
      if (!notebook.matchContext(cwd)) {
        continue;
      }

      Log.debug('Notebook.infer: MATCHED_LISTED_NOTEBOOK %s', notebook.path);
      return notebook;
    }

    Log.debug('Notebook.infer: NO_NOTEBOOK_FOUND');
    return null;
  }

  async function list(cwd: string = process.cwd()): Promise<Notebook[]> {
    Log.debug('list: cwd=%s', cwd);
    const registered_notebooks: Notebook[] = [];

    // STEP 2: Check for notebook configs in config.notebooks
    for (const notebookPath of serviceOptions.config.notebooks) {
      const configFilePath = await Notebook.isNotebookPath(notebookPath);
      if (!configFilePath) {
        continue;
      }
      Log.debug('list.AttemptLoadNotebook: %s', notebookPath);
      const notebook = await Notebook.load(notebookPath);
      if (!notebook) {
        continue;
      }
      registered_notebooks.push(notebook);
    }

    Log.debug('list: found %d notebooks from config', registered_notebooks.length);

    const ancestor_notebooks: Notebook[] = [];
    // Step 3: Search ancestor directories
    let next = cwd;
    while (next !== '/') {
      const configFilePath = await Notebook.isNotebookPath(next);
      const notebookPath = next;
      next = dirname(next);

      if (!configFilePath) {
        continue;
      }

      const notebook = await Notebook.load(notebookPath);
      if (!notebook) {
        continue;
      }

      ancestor_notebooks.push(notebook);
    }
    Log.debug('list: found %d ancestor notebooks', ancestor_notebooks.length);

    return [...registered_notebooks, ...ancestor_notebooks];
  }

  /**
   * Get information about a notebook
   */
  async function info(args: { notebook: Notebook } | { notebookPath: string }) {
    let notebook: Notebook | null = null;
    if ('notebook' in args) {
      notebook = args.notebook;
    } else if ('notebookPath' in args) {
      notebook = await open(args.notebookPath);
    }

    if (!notebook) {
      Log.error('info: NOTEBOOK_NOT_FOUND');
      return null;
    }

    const content = await TuiTemplates.NotebookInfo(notebook);
    // eslint-disable-next-line no-console
    console.log(content);
  }

  Log.debug('Ready');
  /**
   * Return the public API
   */
  return {
    list,
    create,
    open,
    info,
    infer,
  };
}

export type NotebookService = ReturnType<typeof createNotebookService>;
