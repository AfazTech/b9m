<?php

use Exception;

class B9mApi {
    private $baseUrl;
    private $apiKey;

    public function __construct($baseUrl, $apiKey) {
        $this->baseUrl = rtrim($baseUrl, '/');
        $this->apiKey = $apiKey;
    }

    private function request($method, $endpoint, $data = []) {
        $url = $this->baseUrl . '/' . ltrim($endpoint, '/');
        $ch = curl_init($url);
        
        $headers = [
            'Content-Type: application/json',
            'Authorization: Bearer ' . $this->apiKey
        ];
        
        $options = [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_CUSTOMREQUEST => strtoupper($method),
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_SSL_VERIFYPEER => false,
            CURLOPT_SSL_VERIFYHOST => false,
            CURLOPT_TIMEOUT=>3,
        ];
        
        if ($method === 'POST' || $method === 'PUT') {
            $options[CURLOPT_POSTFIELDS] = json_encode($data);
        }
        
        curl_setopt_array($ch, $options);
        
        $response = curl_exec($ch);
        $statusCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);
        
        if ($error) {
           throw new Exception($error);
        }
        
        $decodedResponse = json_decode($response, true);
        if (!$decodedResponse['ok']){
            throw new Exception(message: $decodedResponse['message']);
        }
        return $decodedResponse;
    }

    public function addDomain($domain, $ns1, $ns2) {
        return $this->request('POST', 'domains', [
            'domain' => $domain,
            'ns1' => $ns1,
            'ns2' => $ns2
        ]);
    }

    public function deleteDomain($domain) {
        return $this->request('DELETE', "domains/$domain");
    }

    public function addRecord($domain, $name, $type, $value, $ttl) {
        return $this->request('POST', "domains/$domain/records", [
            'name' => $name,
            'type' => $type,
            'value' => $value,
            'ttl' => $ttl
        ]);
    }

    public function deleteRecord($domain, $name) {
        return $this->request('DELETE', "domains/$domain/records/$name");
    }

    public function getAllRecords($domain) {
        return $this->request('GET', "domains/$domain/records");
    }

    public function reload() {
        return $this->request('POST', 'reload');
    }

    public function restart() {
        return $this->request('POST', 'restart');
    }

    public function stop() {
        return $this->request('POST', 'stop');
    }

    public function start() {
        return $this->request('POST', 'start');
    }

    public function getStatus() {
        return $this->request('GET', 'status');
    }

    public function getStats() {
        return $this->request('GET', 'stats');
    }

    public function getDomains() {
        return $this->request('GET', 'domains');
    }
    
}