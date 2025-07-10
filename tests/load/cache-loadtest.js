/**
 * Copyright 2025 Patryk Rostkowski
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Client, StatusOK } from 'k6/net/grpc';
import { check, sleep } from 'k6';
import encoding from 'k6/encoding';

export const options = {
  vus: 100,
  duration: '60s',
};

function bytesToString(buffer) {
  return String.fromCharCode(...new Uint8Array(buffer));
}

const client = new Client();
client.load(['../../api'], 'cache.proto');

export default () => {
  client.connect('127.0.0.1:8080', { plaintext: true, reflect: true });

  const key = `hello-${__VU}-${__ITER}`;
  const encodedValue = encoding.b64encode(key);

  const setData = { key: key, value: encodedValue };
  const setResponse = client.invoke('cache.CacheService/Set', setData);

  check(setResponse, {
    'set status is OK': (r) => r && r.status === StatusOK,
  });

  console.log(`SET: VU ${__VU}, ITER ${__ITER}, key: ${key}`);

  const getData = { key: key };
  const getResponse = client.invoke('cache.CacheService/Get', getData);

  check(getResponse, {
    'get status is OK': (r) => r && r.status === StatusOK,
    'value was found': (r) => r && r.message && r.message.found === true,
    'value matches': (r) => {
      if (!r || !r.message || !r.message.value) return false;
      const raw = encoding.b64decode(r.message.value);
      return bytesToString(raw) === key;
    },
  });

  const raw = encoding.b64decode(getResponse.message.value);
  console.log(`GET: VU ${__VU}, ITER ${__ITER}, key: ${key}, value: ${bytesToString(raw)}`);

  if (__VU % 5 === 0) {
    const delData = { key: key };
    const delResponse = client.invoke('cache.CacheService/Delete', delData);

    check(delResponse, {
      'delete status is OK': (r) => r && r.status === StatusOK,
    });

    console.log(`DELETE: VU ${__VU}, ITER ${__ITER}, key: ${key}`);
  }

  client.close();
};
