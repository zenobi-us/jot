import { render } from 'binja';
import { Marked } from 'marked';
import TuiRenderer from 'marked-terminal';

const TuiMark = new Marked();
TuiMark.setOptions({
  renderer: new TuiRenderer(),
  gfm: true,
  breaks: false,
});

const TuiRender = async (template: string, ctx: Parameters<typeof render>[1]) => {
  const markdown = await TuiMark.parse(template);
  return render(markdown, ctx);
};

export { TuiRender };
