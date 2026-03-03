import {createContext, useState} from 'react';

export const GlobalContext = createContext();

export const GlobalProvider = ({children}) => {
  const [neoneServerBaseAddress, setNeoneServerBaseAddress] = useState(null);
  const [neoneAuthToken, setNeoneAuthToken] = useState(null);

  return (
    <GlobalContext.Provider
      value={{neoneAuthToken, setNeoneAuthToken, neoneServerBaseAddress, setNeoneServerBaseAddress}}>
      {children}
    </GlobalContext.Provider>
  );
};