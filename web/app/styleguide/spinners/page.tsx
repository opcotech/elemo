import Page from '@/components/Page';
import Spinner from '@/components/Spinner';

export const metadata = {
  title: 'Spinners | Styleguide | Elemo',
  description: 'Styleguide for Elemo'
};

const spinners: string[][] = [
  ['black'],
  ['red-50', 'red-100', 'red-200', 'red-300', 'red-400', 'red-500', 'red-600', 'red-700', 'red-800', 'red-900'],
  [
    'green-50',
    'green-100',
    'green-200',
    'green-300',
    'green-400',
    'green-500',
    'green-600',
    'green-700',
    'green-800',
    'green-900'
  ],
  [
    'blue-50',
    'blue-100',
    'blue-200',
    'blue-300',
    'blue-400',
    'blue-500',
    'blue-600',
    'blue-700',
    'blue-800',
    'blue-900'
  ],
  [
    'orange-50',
    'orange-100',
    'orange-200',
    'orange-300',
    'orange-400',
    'orange-500',
    'orange-600',
    'orange-700',
    'orange-800',
    'orange-900'
  ],
  [
    'yellow-50',
    'yellow-100',
    'yellow-200',
    'yellow-300',
    'yellow-400',
    'yellow-500',
    'yellow-600',
    'yellow-700',
    'yellow-800',
    'yellow-900'
  ],
  [
    'indigo-50',
    'indigo-100',
    'indigo-200',
    'indigo-300',
    'indigo-400',
    'indigo-500',
    'indigo-600',
    'indigo-700',
    'indigo-800',
    'indigo-900'
  ],
  [
    'gray-50',
    'gray-100',
    'gray-200',
    'gray-300',
    'gray-400',
    'gray-500',
    'gray-600',
    'gray-700',
    'gray-800',
    'gray-900'
  ]
];

function SpinnerBox({spinner}: { spinner: string }) {
  return (
    <div className="block">
      <div className={`w-20 h-20 mb-2 text-${spinner}`}>
        <Spinner/>
      </div>
      <div className="text-sm text-center">{spinner}</div>
    </div>
  );
}

function SpinnerRow({spinners}: { spinners: string[] }) {
  return (
    <div className="flex flex-row space-x-4 mb-2">
      {spinners.map((spinner) => (
        <SpinnerBox key={spinner} spinner={spinner}/>
      ))}
    </div>
  );
}

export default function SpinnersPage() {
  return (
    <Page title="Spinners">
      {spinners.map((row, i) => (
        <SpinnerRow key={i} spinners={row}/>
      ))}
    </Page>
  );
}
