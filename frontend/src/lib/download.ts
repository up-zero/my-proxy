export const downLoadRes = (res: any, filename?: string) => {
  // console.log(res)
  // console.log(res.headers.get('Content-Disposition'))
  // let blob=res
console.log(res)
  const headers = res.headers;
  const data = res.data;
  console.log(headers.toString());
  const fileName = /.*filename=(.*)/i.exec(headers.get("Content-Disposition"))?.[1] || filename || "下载.xls";
  // 开始下载

  const blob = new Blob([data], { type: headers["content-type"] });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};
export const downloadJson=(data: any, fileName = 'data.json')=> {
  const jsonStr = JSON.stringify(data, null, 2)
  const blob = new Blob([jsonStr], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)

  const a = document.createElement('a')
  a.href = url
  a.download = fileName
  a.click()

  URL.revokeObjectURL(url)
}
export const downLoadUrl = (url: string, name: string) => {
  const link = document.createElement("a");
  link.href = url;
  link.download = name;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};
export const downLoadUrl2 = (url: string, name: string) => {
  fetch(url)
  .then(res => res.blob())
  .then(blob => {
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = name
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  })
};
