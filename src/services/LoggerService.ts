import { pino } from 'pino';
import pinoPretty from 'pino-pretty';

export const Logger = pino(
  {
    name: 'wiki',
    level: process.env.LOG_LEVEL || 'info',
    enabled: !!process.env.DEBUG,
  },
  pinoPretty({
    colorize: true,
    singleLine: true,
    messageFormat: '{name}{if namespace}.{namespace}{end} {msg}',
  })
);
