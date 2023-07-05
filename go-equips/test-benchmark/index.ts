import fetch from 'cross-fetch';

let startTime: any, endTime: any;

const start = () => {
  startTime = new Date();
}

const end = () => {
  endTime = new Date();
  return Math.round(endTime - startTime)
};

function calculateAverage(numbers: any) {
    const sum = numbers.reduce((total: any, num: any) => total + num, 0);
    const average = sum / numbers.length;
    return average;
}

(async () => {
    const times: any = []
    for(let i = 0; i < 100; i++){
        start()
        const res = await fetch('http://localhost:7077/all')
        times.push(end())
    }
    console.log(calculateAverage(times))
})()    