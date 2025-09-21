import config from "@/config/index";

export enum ILogLevel {
  trace = 1,
  debug = 2,
  info = 3,
  warn = 4,
  error = 5,
}

function log(type: ILogLevel) {
  return (name: string, ...args: any[]) => {
    // 日志等级小于配置等级则不输出
    const configLogNum = Number(ILogLevel[config.logLevel as any]);
    if (type < configLogNum) return;
    // 根据等级输出不同颜色
    if (type === 1) {
      console.log(`%c ${name}`, "background:#40a9ff;color:#fff;padding:3px;", ...args);
    } else if (type === 2) {
      console.log(`%c ${name}`, "background:#002766;color:#fff;padding:3px;", ...args);
    } else if (type === 3) {
      console.log(`%c ${name}`, "background:#006633;color:#fff;padding:3px;", ...args);
    } else if (type === 4) {
      console.log(`%c ${name}`, "background:#FF9966;color:#fff;padding:3px;", ...args);
    } else if (type === 5) {
      console.log(`%c ${name}`, "background:#CC3333;color:#fff;padding:3px;", ...args);
    }
  };
}

export default {
  trace: log(1),
  debug: log(2),
  info: log(3),
  warn: log(4),
  error: log(5),
};
